package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	validStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	invalidStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
)

type tool struct {
	name, nixPkg string
	selected     bool
}

// Struct for the input JSON file
type pathConfig struct {
	Hidden   []string `json:"hidden"`
	ReadOnly []string `json:"read-only"`
}

type model struct {
	tools           []tool
	cursor          int
	network         bool
	configPathInput textinput.Model
	configData      pathConfig
	hiddenPathInput textinput.Model
	step            int // 0: Tools, 1: Network, 2: JSON Path, 3: Hidden Paths
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "paths.json"
	ti.Focus()

	home, _ := os.UserHomeDir()
	ti2 := textinput.New()
	// Default suggested hidden paths
	ti2.SetValue(fmt.Sprintf("%s/.ssh, %s/.aws", home, home))

	return model{
		tools: []tool{
			{name: "Node.js (LTS)", nixPkg: "nodejs", selected: true},
			{name: "Python", nixPkg: "python3", selected: true},
			{name: "TypeScript", nixPkg: "typescript", selected: false},
			{name: "Go", nixPkg: "go", selected: false},
			{name: "Rust", nixPkg: "rust", selected: false},
			{name: "Bun", nixPkg: "bun", selected: false},
		},
		network:         true,
		configPathInput: ti,
		hiddenPathInput: ti2,
		step:            0,
	}
}

// autoCompletePath attempts to find a file matching the current prefix
func autoCompletePath(path string) string {
	if path == "" {
		return path
	}
	expanded := path
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		expanded = filepath.Join(home, path[1:])
	}
	matches, err := filepath.Glob(expanded + "*")
	if err == nil && len(matches) > 0 {
		// If the user started with tilde, try to return a tilde-based completion
		if strings.HasPrefix(path, "~") {
			home, _ := os.UserHomeDir()
			rel, err := filepath.Rel(home, matches[0])
			if err == nil {
				return "~/" + rel
			}
		}
		return matches[0]
	}
	return path
}

func checkPathStyled(p string) string {
	expanded := p
	if strings.HasPrefix(p, "~") {
		home, _ := os.UserHomeDir()
		expanded = filepath.Join(home, p[1:])
	}
	if _, err := os.Stat(expanded); err == nil {
		return validStyle.Render("  ✔ " + p)
	}
	return invalidStyle.Render("  ✘ " + p + " (missing)")
}

func validateJSONConfig(path string) (pathConfig, string) {
	var cfg pathConfig
	expanded := path
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		expanded = filepath.Join(home, path[1:])
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		return cfg, "File not found"
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, "Invalid JSON format"
	}
	return cfg, "ok"
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			if m.step == 2 {
				m.configPathInput.SetValue(autoCompletePath(m.configPathInput.Value()))
				m.configPathInput.SetCursor(len(m.configPathInput.Value()))
			}
		case "up", "k":
			if m.step == 0 && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.step == 0 && m.cursor < len(m.tools)-1 {
				m.cursor++
			}
		case " ":
			if m.step == 0 {
				m.tools[m.cursor].selected = !m.tools[m.cursor].selected
			}
		case "enter":
			switch m.step {
			case 0, 1:
				m.step++
			case 2:
				cfg, status := validateJSONConfig(m.configPathInput.Value())
				if status == "ok" {
					m.configData = cfg
					m.step++
					m.hiddenPathInput.Focus()
				}
			case 3:
				m.saveAndExit()
				return m, tea.Quit
			}
		case "y", "n":
			if m.step == 1 {
				m.network = (msg.String() == "y")
				m.step++
			}
		case "<":
			if m.step > 0 {
				m.step--
			}
		}
	}

	var cmd tea.Cmd
	if m.step == 2 {
		m.configPathInput, cmd = m.configPathInput.Update(msg)
	} else if m.step == 3 {
		m.hiddenPathInput, cmd = m.hiddenPathInput.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	s := titleStyle.Render("📦 OpenBox Configurator") + "\n\n"

	switch m.step {
	case 0:
		s += "Step 1: Select Tools (Space to toggle)\n\n"
		for i, t := range m.tools {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := " "
			if t.selected {
				checked = "x"
			}
			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, t.name)
		}
	case 1:
		s += "Step 2: Network Permissions\n\n"
		s += fmt.Sprintf("Allow Internet Access? (y/n): %v\n", m.network)
	case 2:
		s += "Step 3: Path Config JSON\n"
		s += helpStyle.Render("Press [Tab] to autocomplete file path\n\n")
		s += m.configPathInput.View() + "\n"
		cfg, status := validateJSONConfig(m.configPathInput.Value())
		if status != "ok" {
			s += invalidStyle.Render("\nStatus: " + status)
		} else {
			s += "\n" + labelStyle.Render("Read-Only Paths Found:") + "\n"
			if len(cfg.ReadOnly) == 0 {
				s += "  (None)\n"
			}
			for _, p := range cfg.ReadOnly {
				s += checkPathStyled(p) + "\n"
			}
		}
	case 3:
		s += "Step 4: Ghost Paths (Completely Invisible)\n"
		s += helpStyle.Render("Separate multiple paths with commas\n\n")
		s += m.hiddenPathInput.View() + "\n\n"
		s += labelStyle.Render("Validated Hidden Paths:") + "\n"
		paths := strings.Split(m.hiddenPathInput.Value(), ",")
		for _, p := range paths {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				s += checkPathStyled(trimmed) + "\n"
			}
		}
	}

	s += helpStyle.Render("\n\n[Enter] Next • [Tab] Auto-path • [<] Previous • [Esc] Quit")
	return s
}

func (m model) saveAndExit() {
	// Process file paths for RW, RO, and hidden directories
	pathMap := make(map[string]string)
	for _, p := range m.configData.ReadOnly {
		pathMap[p] = "ro"
	}

	var hiddenList []string
	for _, p := range m.configData.Hidden {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			hiddenList = append(hiddenList, trimmed)
		}
	}

	prefs := map[string]interface{}{
		"network":      m.network,
		"extra_paths":  pathMap,
		"hidden_paths": hiddenList,
	}
	file, _ := json.MarshalIndent(prefs, "", "  ")
	_ = os.WriteFile(".sandbox_prefs", file, 0644)

	// Build Package List
	var pkgs []string
	for _, t := range m.tools {
		if t.selected {
			pkgs = append(pkgs, t.nixPkg)
		}
	}
	allPkgs := append([]string{"bash-completion", "fzf", "vim", "git"}, pkgs...)

	setupScript := fmt.Sprintf(`#!/bin/bash
WSP="${WORKSPACE_PATH:-$(pwd)/workspace}"
export REAL_HOME_PATH="%s"
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
`, os.Getenv("HOME"))

	_ = os.WriteFile("setup-env.sh", []byte(setupScript), 0755)

	// 4. Generate flake.nix
	flake := fmt.Sprintf(`{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in {
      devShells = forAllSystems (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ %s ];
            shellHook = "source ./setup-env.sh";
          };
        });
    };
}`, strings.Join(allPkgs, " "))

	_ = os.WriteFile("flake.nix", []byte(flake), 0644)

	// Final exit message/logic
	fmt.Println("Config saved. Run your sandbox script to enter.")
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
