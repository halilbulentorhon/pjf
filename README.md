# ProjectFinder (pjf)

> A fast terminal UI for discovering, organizing, and managing your local development projects.

pjf scans your filesystem for git repositories, presents them in a fuzzy-searchable grouped list, and lets you open them in your IDE or terminal, run commands, and manage project organization — all without leaving the terminal.

## Install

```bash
curl -sSL https://raw.githubusercontent.com/halilbulentorhon/pjf/main/install.sh | sh
```

**Update:** `pjf update` | **Uninstall:** `pjf uninstall`

## Quick Start

Run `pjf`. On first launch, a wizard asks which directories to scan. After that, you're in.

## What Can It Do?

- **Fuzzy search** across hundreds of projects instantly
- **Organize into groups** — collapsible, reorderable (Work, Personal, etc.)
- **Open in IDE** — auto-detects VS Code, Cursor, IntelliJ, GoLand, WebStorm, Zed, Claude Code, Neovim, Vim
- **Smart IDE defaults** — Go → GoLand, Node → VS Code, override per project
- **Run commands** — global and per-project, with output viewer
- **Manage everything in-app** — press `s` for settings, no YAML editing needed

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `enter` | Action menu |
| `t` | Open in terminal |
| `o` | Open in default IDE |
| `m` | Move to group |
| `s` | Settings |
| `r` | Rescan |
| `h` | Toggle hidden |
| `?` | Help |
| `q` | Quit |
| `←` / `→` | Collapse/expand groups |
| `u` / `d` | Reorder groups |
| `esc` | Search bar |

Search is focused by default — just type. `↓` to list, `esc` to clear. Menus: `1-9` to pick by number.

## Platform Support

macOS only for now. Linux and Windows planned.

## License

MIT

---

<sub>Vibe coded with [Claude Code](https://claude.ai/code) from start to finish.</sub>
