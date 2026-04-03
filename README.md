# ProjectFinder (pjf)

> A fast terminal UI for discovering, organizing, and managing your local development projects.

pjf scans your filesystem for git repositories, presents them in a fuzzy-searchable grouped list, and lets you open them in your IDE or terminal, run commands, and manage project organization — all without leaving the terminal.

<!-- TODO: Add screenshot/GIF here after first release -->

## Install

```bash
curl -sSL https://raw.githubusercontent.com/halilbulentorhon/pjf/main/install.sh | sh
```

<details>
<summary>Other installation methods</summary>

**go install:**
```bash
go install github.com/halilbulentorhon/pjf@latest
```

**From source:**
```bash
git clone https://github.com/halilbulentorhon/pjf.git
cd pjf
go build -o pjf .
mv pjf /usr/local/bin/
```
</details>

**Uninstall:** `pjf uninstall`

## Quick Start

```bash
pjf
```

On first run, a setup wizard asks which directories to scan. After scanning, you'll see your projects listed and ready to go.

## What Can It Do?

**Find projects fast** — Fuzzy search across hundreds of projects instantly.

**Organize into groups** — Create groups like Work, Personal, Open Source. Collapse/expand with arrow keys.

**Open anywhere** — Press `t` for terminal, `o` for your IDE. pjf auto-detects VS Code, Cursor, IntelliJ, GoLand, WebStorm, Zed, Claude Code, Neovim, and Vim.

**Smart IDE defaults** — Set Go projects to open in GoLand, Node in VS Code. Override per project.

**Run commands** — Save frequently used commands (build, test, deploy) per project or globally. Run ad-hoc commands with output viewer.

**Manage everything in-app** — Press `s` for settings. No YAML editing required.

## Keyboard Shortcuts

### List View

| Key | Action |
|-----|--------|
| `enter` | Action menu |
| `t` | Open in terminal |
| `o` | Open in default IDE |
| `m` | Move to group |
| `s` | Settings |
| `r` | Rescan projects |
| `h` | Toggle hidden projects |
| `?` | Help |
| `q` | Quit |
| `←` / `→` | Collapse/expand groups |
| `u` / `d` | Reorder groups |
| `esc` | Jump to search bar |

### Search

Search is focused by default — just start typing. Press `↓` to go to the list. Press `esc` to clear (or quit if empty).

### Menus

`1-9` select by number. `↑/↓` navigate. `enter` confirm. `esc` back.

## Action Menu

Press `enter` on any project:

| # | Action | Description |
|---|--------|-------------|
| 1 | Open in IDE | Pick from detected IDEs, set defaults |
| 2 | Open in Terminal | New terminal window at project dir |
| 3 | Run Command | Saved or ad-hoc commands with output |
| 4 | Copy Path | Copy to clipboard |
| 5 | Project Settings | IDE override, saved commands |
| 6 | Add to Group | Organize into groups |
| 7 | Hide from List | Reversible |
| 8 | Delete Project | With confirmation |

## Configuration

Config: `~/.config/pjf/config.yaml` — most settings manageable from the TUI (press `s`).

<details>
<summary>Full config reference</summary>

```yaml
scan_dirs:
  - ~/Source
  - ~/work

extra_excludes:
  - ~/Source/archived

default_ides:
  go: goland
  node: code
  _default: cursor

project_ides:
  ~/Source/special-project: zed

global_commands:
  - name: Git Status
    command: git status
  - name: Docker Up
    command: docker compose up -d --build

project_commands:
  - path: ~/work/backend-api
    commands:
      - name: Run Tests
        command: go test ./...

groups:
  - name: Work
    collapsed: false
    projects:
      - ~/work/backend-api
      - ~/work/frontend

hidden_projects: []
max_depth: 5
cache_ttl: 24
```
</details>

### IDE Detection

pjf auto-detects IDEs from PATH (`code`, `cursor`, `idea`, `goland`, `webstorm`, `zed`, `claude`, `nvim`, `vim`) and macOS application directories (`/Applications/`, JetBrains Toolbox).

Set defaults per project type with `default_ides`, or override per project with `project_ides`.

## Platform Support

| Platform | Status |
|----------|--------|
| macOS (Apple Silicon) | Supported |
| macOS (Intel) | Supported |
| Linux | Planned |
| Windows | Planned |

## Contributing

Issues and pull requests welcome.

## License

MIT
