# ProjectFinder (pjf)

A fast terminal UI for discovering, organizing, and managing your local development projects.

pjf scans your filesystem for git repositories, presents them in a fuzzy-searchable grouped list, and lets you open them in your IDE or terminal, run commands, and manage project organization — all without leaving the terminal.

## Features

- **Auto-discovery** — Scans directories for git repos with smart excludes
- **Fuzzy search** — Type to instantly filter across all projects
- **Project groups** — Organize projects into collapsible groups (Work, Personal, etc.)
- **IDE integration** — Auto-detects installed IDEs (VS Code, Cursor, IntelliJ, GoLand, WebStorm, Zed, Claude Code, Neovim, Vim)
- **Type-based IDE defaults** — Set default IDE per project type (Go → GoLand, Node → VS Code)
- **Custom commands** — Global and per-project saved commands with output viewer
- **Quick actions** — Single-key shortcuts for common operations
- **Settings UI** — Manage all configuration from within the TUI
- **Smart caching** — Instant startup with background refresh

## Installation

### From source

Requires [Go 1.25+](https://go.dev/dl/).

```bash
git clone https://github.com/halilbulentorhon/pjf.git
cd pjf
go build -o pjf .
sudo mv pjf /usr/local/bin/
```

### Uninstall

```bash
pjf uninstall
```

Removes config (`~/.config/pjf/`), cache (`~/.cache/pjf/`), and the binary.

## Quick Start

```bash
pjf
```

On first run, a setup wizard asks which directories to scan. After scanning, you'll see your projects listed.

## Keyboard Shortcuts

### List View

| Key | Action |
|-----|--------|
| Any character | Fuzzy search (when search focused) |
| `↑` / `↓` | Navigate list / search |
| `enter` | Open action menu |
| `t` | Open in terminal |
| `o` | Open in default IDE |
| `m` | Move to group |
| `r` | Rescan projects |
| `h` | Toggle hidden projects |
| `s` | Settings |
| `?` | Help |
| `q` | Quit |
| `←` | Collapse group / jump to header |
| `→` | Expand group / jump to next group |
| `u` / `d` | Reorder groups (on header) |
| `esc` | Search → clear, List → search, Empty search → quit |

### Menus

| Key | Action |
|-----|--------|
| `1-9` | Select item by number |
| `↑` / `↓` | Navigate |
| `enter` | Confirm |
| `esc` | Back |

### Search Mode

Search bar is focused by default. Start typing to filter. Press `↓` to move to the list and enable shortcuts. Press `esc` or `↑` at top to return to search.

## Action Menu

Select a project and press `enter`:

1. **Open in IDE** — Pick from detected IDEs, set defaults
2. **Open in Terminal** — Opens a new terminal window at the project
3. **Run Command** — Saved commands or ad-hoc input with output viewer
4. **Copy Path** — Copies project path to clipboard
5. **Project Settings** — Manage IDE override and saved commands
6. **Add to Group** — Assign to a group or create new
7. **Hide from List** — Hide project (reversible)
8. **Delete Project** — Remove from filesystem (with confirmation)

## Configuration

Config file: `~/.config/pjf/config.yaml`

Most settings can be managed from the TUI (press `s`), but here's the full config for reference:

```yaml
# Directories to scan for projects
scan_dirs:
  - ~/Source
  - ~/work

# Additional directories to exclude from scanning
extra_excludes:
  - ~/Source/archived

# Default IDE per project type (go, node, java, rust, python, unknown)
default_ides:
  go: goland
  node: code
  _default: cursor

# Per-project IDE override
project_ides:
  ~/Source/special-project: zed

# Global commands (available for all projects)
global_commands:
  - name: Git Status
    command: git status
  - name: Docker Up
    command: docker compose up -d --build

# Per-project commands
project_commands:
  - path: ~/work/backend-api
    commands:
      - name: Run Tests
        command: go test ./...
      - name: Build
        command: go build ./...

# Project groups
groups:
  - name: Work
    collapsed: false
    projects:
      - ~/work/backend-api
      - ~/work/frontend
  - name: Personal
    collapsed: false
    projects:
      - ~/Source/my-blog

# Hidden projects (managed via TUI)
hidden_projects: []

# Scan depth (default: 5)
max_depth: 5

# Cache TTL in hours (default: 24)
cache_ttl: 24
```

### IDE Detection

pjf auto-detects IDEs from:
- **PATH**: `code`, `cursor`, `idea`, `goland`, `webstorm`, `zed`, `claude`, `nvim`, `vim`
- **macOS Apps**: `/Applications/`, JetBrains Toolbox paths

Set a default IDE per project type with `default_ides`, or override per project with `project_ides`.

## Platform Support

- **macOS** — Full support
- **Linux** — Planned
- **Windows** — Planned

## Current Limitations

- macOS only (Linux and Windows support planned)
- No `_linux.go` implementations yet for terminal opener and clipboard
- No plugin system for custom IDE definitions (use config overrides)
- Group ordering is manual only (no auto-sort)

## License

MIT

## Contributing

Issues and pull requests welcome at [github.com/halilbulentorhon/pjf](https://github.com/halilbulentorhon/pjf).
