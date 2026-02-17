# 📦 OpenBox

**The ultimate zero-trust sandbox for autonomous AI agents.**

**OpenBox** is a high-security, TUI-driven sandbox designed to run AI agents (like OpenClaw) without risking your personal data. By combining macOS kernel-level `sandbox-exec` with Nix package management, it creates a "jail" that ensures what happens in the box, stays in the box.

## ✨ Why OpenBox?

As AI agents become more autonomous, giving them shell access is a significant security risk. **OpenBox** solves the "Who watches the watcher?" problem:

- **Zero-Trust Environment:** No longer worry about an AI agent recursively deleting your home directory or exfiltrating your browser cookies.
- **JSON-Defined Scope:** Use a standardized `paths.json` to specify exactly which directories the agent can "see" (Read-Only) or "touch" (Read-Write).
- **Ghosting (Invisible Paths):** Define highly sensitive directories in your configuration to be made **completely invisible** to the agent via kernel-level denial.
- **Agent Framework Ready:** Built-in support for **OpenClaw**, allowing you to install and launch gateway services within the isolated environment instantly.

## 🛠️ Key Features

- **Kernel-Level Enforcement:** Uses macOS "Seatbelt" profiles to block file-system access at the kernel level.
- **Language-Agnostic Isolation:** Automatically redirects caches and toolchains for **npm, pip, Go, and Cargo** into the local project `./workspace`.
- **Zsh & Powerlevel10k Integration:** Sourced directly from your system for a familiar, high-performance terminal experience, but re-locked to the sandbox `$HOME`.
- **Reproducible Tools:** Uses Nix Flakes to provide the agent with exact tool versions, independent of your system's global configuration.

---

## 🚀 Installation & Setup

### Prerequisites

- **macOS** (Optimized for Sequoia/Sonoma)
- **Nix Package Manager** (with Flakes enabled)
- **Go** (For the configuration TUI)

### 0. Configure Your Directory

Create the initial file structure and check for requirements with

```bash
./init.sh
```

### 1. Configure Your Box

OpenBox uses an interactive TUI to generate your security profile. It reads from a JSON file to populate whitelisted paths.

```bash
./run-sandbox.sh

```

- **Tools:** Select Node.js, Python, Go, Rust, etc.
- **JSON Config:** Point the TUI to your `paths.json` for Read-Only/Read-Write rules.
- **Ghosting:** Specify comma-separated paths to hide entirely.
- **Speed:** Automatically generates a `.gitignore` to keep `node_modules` out of the Nix store for instant startup.

### 📄 `paths.json` Template

To make the most of **OpenBox**, use this template. It covers the most common scenarios: allowing an agent to read your documentation, write to a specific project folder, but stay away from your private keys.

```json
{
  "read-write": [
    "~/Documents/ai-projects/active-agent-workspace",
    "/tmp/openbox-logs"
  ],
  "read-only": [
    "~/Documents/reference-docs",
    "~/.config/git",
    "/usr/local/bin"
  ],
  "hidden": ["~/.ssh", "~/.aws", "~/.env", "~/Library/Keychains"]
}
```

### 2. Enter the Sandbox

After setup with the TUI, it should continue into the sandbox. For subsequent uses, you can rerun the same command.

```bash
./run-sandbox.sh

```

### 3. Deploy OpenClaw

Once inside the 📦 environment, you have dedicated commands to manage your AI gateway. Note that openclaw-install will go through the OpenClaw onboarding, but it **will** produce errors when it comes to initial setup of the gateway. This is **GOOD** since it means the agent can't get through the sandbox any way we don't specify. The `openclaw-launch` opens this gateway manually.

```bash
# Install OpenClaw in the isolated workspace
openclaw-install

# Start the OpenClaw Gateway service
openclaw-launch

```

---

## 🏗️ Architecture

OpenBox operates on a layered security model:

1. **Nix Layer:** Provides a pure environment where tools are symlinked, not installed globally.
2. **Environment Layer:** Redirects all `HOME` style variables (`$CARGO_HOME`, `$npm_config_cache`, etc.) to the local `./workspace`.
3. **Kernel Layer:** The `sandbox-exec` engine (or `bwrap` on Linux) denies any syscall attempting to read/write outside of permitted paths defined in your JSON and TUI setup.

---

## ⚠️ Known Limitations

- **GUI Applications:** Designed for CLI-based agents. Launching macOS GUI apps (like VS Code) may fail due to restricted access to system services.
- **Sudo:** `sudo` privileges are disabled within the box. Agents cannot perform administrative actions.
- **Symlink Resolution:** Some host symlinks may require their _absolute real path_ to be explicitly whitelisted in your JSON config.

### 🐧 Linux Support

## The Linux implementation is planned to use **Bubblewrap (`bwrap`)**, but it not yet developed. Help is appreciated

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---
