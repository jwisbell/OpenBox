#!/bin/bash

# Exit on error
set -e

echo "📦  Initializing OpenBox Development Environment..."

# 1. Ensure Nix is installed
if ! command -v nix &>/dev/null; then
  echo "❌ Nix not found. Please install it from https://nixos.org/download.html"
  exit 1
fi

# 2. Enable Experimental Features (Local to this session if not in config)
# This prevents the 'nix-command' error for the rest of the script
export NIX_CONFIG="extra-experimental-features = nix-command flakes"

# 3. Create necessary directories
echo "📁 Creating workspace and local cache..."
mkdir -p workspace
touch workspace/.viminfo workspace/.bash_history workspace/.zsh_history

# 4. Initialize Go modules (if not already done)
if [ ! -f "go.mod" ]; then
  echo "🐹 Initializing Go modules..."
  go mod init sandbox-setup
  go mod tidy
fi

# 5. Fetch dependencies for the TUI
echo "📦 Fetching TUI dependencies..."
go get github.com/charmbracelet/bubbles/textinput
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss

# 6. Make scripts executable
chmod +x run-sandbox.sh

echo -e "\n✅ Setup Complete!"
echo "🚀 Run 'go run cmd/setup/main.go' to configure your sandbox."
echo "🛠️  Then run './run-sandbox.sh' to enter."
