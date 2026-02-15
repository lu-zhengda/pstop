# pstop

Process explorer for macOS -- browse, search, and manage processes with a live-updating TUI or handy CLI subcommands.

## Install

```bash
brew tap lu-zhengda/tap
brew install pstop
```

## Quick Start

```bash
pstop           # Launch interactive TUI
pstop --help    # Show all commands
```

## Commands

| Command              | Description                                    |
|----------------------|------------------------------------------------|
| `list`               | List all processes (--sort cpu\|mem\|pid\|name) |
| `top`                | Top resource consumers (--n 10, --battery)     |
| `find <query>`       | Find by name, command, or port                 |
| `info <pid>`         | Detailed process info (files, ports, children) |
| `kill <pid>`         | Kill process (--force, --signal SIG)           |
| `tree`               | Process tree view                              |
| `dev`                | Developer view grouped by stack                |
| `watch <pid>`        | Live-monitor a process (--interval 2)          |

## TUI

Launch without arguments for interactive mode. Features:

- Live-updating process table (refreshes every 2s)
- Sort by CPU, MEM, PID, or Name (press 1-4)
- Search/filter with `/`
- Tab switching: All | Top | Dev (Tab key)
- Kill selected process with `K` (with confirmation)
- View detailed info with `i`
- Navigate with `j`/`k`, page with PgUp/PgDn

Color coding: red for high CPU (>50%), yellow for medium (20-50%).

## License

MIT
