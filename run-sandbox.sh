#!/bin/bash

# 1. Check if we need to run the setup
if [ ! -f "flake.nix" ]; then
  echo "First time setup: Building Go TUI..."
  nix --extra-experimental-features "nix-command flakes" run nixpkgs#go -- run cmd/setup/main.go
fi

# 2. Extract Prefs
NETWORK=$(python3 -c "import json; print(json.load(open('.sandbox_prefs'))['network'])" 2>/dev/null || echo "True")
PATH_DATA=$(python3 -c "import json; d=json.load(open('.sandbox_prefs'))['extra_paths']; print(' '.join([f'{k},{v}' for k,v in d.items()]))" 2>/dev/null)
HIDDEN_PATHS=$(python3 -c "import json; print(' '.join(json.load(open('.sandbox_prefs'))['hidden_paths']))" 2>/dev/null)

mkdir -p ./workspace

# 3. Launch Sandbox (Linux or Mac)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  # Linux (bwrap) is already deny-by-default (unshare-all)
  NET_FLAG=$([ "$NETWORK" == "True" ] && echo "--share-net" || echo "--unshare-net")
  BWRAP_FLAGS="--dev-bind / / --ro-bind /nix /nix --bind ./workspace /app"

  for item in $PATH_DATA; do
    IFS=',' read -r path mode <<<"$item"
    if [ -e "$path" ]; then
      BWRAP_FLAGS="$BWRAP_FLAGS $([ "$mode" == "ro" ] && echo "--ro-bind" || echo "--bind") $path $path"
    fi
  done

  bwrap $BWRAP_FLAGS $NET_FLAG nix --extra-experimental-features "nix-command flakes" develop --command bash

elif [[ "$OSTYPE" == "darwin"* ]]; then
  # Load environment variables if present
  [[ -f .env ]] && export $(grep -v '^#' .env | xargs)

  # 1. Setup Path Variables
  WORKSPACE_ABS="$(pwd)/workspace"
  NIX_CACHE="$HOME/.cache/nix"
  P10K_CONFIG="$HOME/.p10k.zsh"
  ZSH_DIR="$HOME/.oh-my-zsh"

  # 2. Configure Network Rule
  if [[ "$NETWORK" == "True" ]]; then
    NET_RULE="(allow network*)"
  else
    NET_RULE="(deny network*)
              (allow network-outbound (remote ip \"127.0.0.1:*\"))
              (allow network-inbound (local ip \"127.0.0.1:*\"))"
  fi

  # 3. Build Base macOS Sandbox Profile
  cat <<EOF >.mac.sb
(version 1)
(allow default)

;; --- ISOLATION LAYER ---
(deny file-write* (subpath "/"))

;; --- WRITE WHITELIST ---
(allow file-write*
  (subpath "/nix")
  (subpath "/private/tmp")
  (subpath "/tmp")
  (subpath "/private/var/folders")
  (subpath "$NIX_CACHE")
  (subpath "$(pwd)")
  (subpath "$WORKSPACE_ABS"))

;; --- READ WHITELIST (Themes & Shell) ---
(allow file-read*
  (subpath "$HOME/.zshrc")
  (subpath "$P10K_CONFIG")
  (subpath "$ZSH_DIR"))

;; --- SYSTEM & DEVICES ---
$NET_RULE
(allow mach-lookup (global-name "com.steipete.openclaw.gateway"))
(allow file-read* file-write*
  (subpath "/dev/null")
  (subpath "/dev/stdin")
  (subpath "/dev/stdout")
  (subpath "/dev/stderr")
  (subpath "/dev/tty"))
EOF

  # 4. Append EXTRA PATHS (Dynamic Whitelist)
  for item in $PATH_DATA; do
    IFS=',' read -r path mode <<<"$item"
    [[ -e "$path" ]] || continue
    if [[ "$mode" == "rw" ]]; then
      echo "(allow file-read* file-write* (subpath \"$path\"))" >>.mac.sb
    else
      echo "(allow file-read* (subpath \"$path\"))" >>.mac.sb
    fi
  done

  # 5. Append GHOST PATHS (Explicit Deny)
  for path in $HIDDEN_PATHS; do
    expanded_path=$(eval echo "$path")
    if [[ -e "$expanded_path" ]]; then
      echo "(deny file* (subpath \"$expanded_path\"))" >>.mac.sb
      echo "(deny file-read-metadata (subpath \"$expanded_path\"))" >>.mac.sb
    fi
  done

  # 6. Final Prep & Launch
  touch "$WORKSPACE_ABS"/.{viminfo,bash_history,zsh_history}

  export WORKSPACE_PATH="$WORKSPACE_ABS"
  export SANDBOX_SHELL="zsh"
  export REAL_ZSHRC="$HOME/.zshrc"
  export REAL_P10K="$P10K_CONFIG"

  echo -e "\033[1;34m📦 Only specified JSON paths and Nix/System tools are visible.\033[0m"

  trap 'rm -f .mac.sb' EXIT
  sandbox-exec -f .mac.sb nix --extra-experimental-features "nix-command flakes" develop --command "$SANDBOX_SHELL"
fi
