#!/bin/bash
WSP="${WORKSPACE_PATH:-$(pwd)/workspace}"
export REAL_HOME_PATH="/Users/jwisbell"
export HOME="$WSP"

# --- PERSISTENCE & ISOLATION ---
export HISTFILE="$WSP/.zsh_history"
export VIMINIT="set noswapfile | set noundofile | set viminfo='100,<50,s10,h,n$WSP/.viminfo"
export OPENCLAW_DIR="$WSP/.openclaw"

# --- LANGUAGE RUNTIME CACHES ---
mkdir -p "$WSP"/{.npm-cache,.npm-global,.python-cache,.go-cache,.cargo,bin,.openclaw}
export npm_config_cache="$WSP/.npm-cache"
export npm_config_prefix="$WSP/.npm-global"
export PIP_CACHE_DIR="$WSP/.python-cache"
export GOMODCACHE="$WSP/.go-cache/mod"
export CARGO_HOME="$WSP/.cargo"
export PATH="$WSP/bin:$WSP/.npm-global/bin:$WSP/.go-path/bin:$WSP/.cargo/bin:$PATH"

# --- OPENCLAW HELPERS ---
cat << 'EOF' > "$WSP/bin/openclaw-install"
#!/bin/bash
export HOME="${WORKSPACE_PATH:-$(pwd)/workspace}"
export OPENCLAW_DIR="$HOME/.openclaw"
echo "📦 Installing OpenClaw inside this container..."
if [ ! -d "node_modules/openclaw" ]; then
    echo "Downloading OpenClaw via NPM..."
    npm install openclaw 
else
    echo "✅ OpenClaw already present in node_modules."
fi

echo "⚙️  Running Onboarding... (If it fails at the network step, that's okay!)"
./node_modules/.bin/openclaw onboard
echo -e "\n✨ Installation complete!"
echo "🚀 To start the service, run: openclaw-launch"
EOF

cat << 'EOF' > "$WSP/bin/openclaw-launch"
#!/bin/bash
WSP_INNER="${WORKSPACE_PATH:-$(pwd)/workspace}"
OC_BIN="$(pwd)/node_modules/.bin/openclaw"
if [ ! -f "$OC_BIN" ]; then
    echo "❌ OpenClaw not found. Run: npm install openclaw"
    exit 1
fi
# Function to kill the process on exit
cleanup() {
    if [ -f "$WSP_INNER/openclaw.pid" ]; then
        PID=$(cat "$WSP_INNER/openclaw.pid")
        echo -e "\n🛑 Shutting down OpenClaw Gateway (PID: $PID)..."
        kill $PID 2>/dev/null
        rm "$WSP_INNER/openclaw.pid"
    fi
}
# Trap exit signals (Ctrl+C, terminal close, etc.)
trap cleanup EXIT
echo "🦀 Starting OpenClaw Gateway (from within OpenBox)..."
$OC_BIN gateway run > "$WSP_INNER/openclaw.log" 2>&1 &
echo $! > "$WSP_INNER/openclaw.pid"
echo "✅ Gateway active. OpenClaw process running."
echo "📝 Logs: tail -f workspace/openclaw.log"
wait

EOF

chmod +x "$WSP/bin/openclaw-install" "$WSP/bin/openclaw-launch"

# --- ZSH / POWERLEVEL10K BRIDGE ---
cat << 'ZSHRC' > "$WSP/.zshrc"
export CLICOLOR=1
export LSCOLORS=Gxfxcxdxbxegedabagacad
autoload -U colors && colors

alias ls="${OSTYPE/darwin*/ls -G}"
[[ "$OSTYPE" != "darwin"* ]] && alias ls='ls --color=auto'
alias l='ls -lah' ll='ls -lh'

# Bridge to real system config
export ZSH="$REAL_HOME_PATH/.oh-my-zsh"
[ -f "$REAL_HOME_PATH/.zshrc" ] && source "$REAL_HOME_PATH/.zshrc"
[ -f "$REAL_HOME_PATH/.p10k.zsh" ] && source "$REAL_HOME_PATH/.p10k.zsh"

# Re-lock environment
export HOME="$WORKSPACE_PATH"
ZSHRC

# --- EXECUTION ---
echo -e "\033[1;32m📦 OpenBox Ready!\033[0m"
if command -v zsh >/dev/null; then
    exec zsh -is eval "export HISTFILE=$WSP/.zsh_history"
fi
