# pstop

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: macOS](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://github.com/lu-zhengda/pstop)
[![Homebrew](https://img.shields.io/badge/Homebrew-lu--zhengda/tap-orange.svg)](https://github.com/lu-zhengda/homebrew-tap)

Process explorer for macOS — browse, search, and manage processes with a live-updating TUI.

## Install

```bash
brew tap lu-zhengda/tap
brew install pstop
```

## Usage

```
$ pstop top --n 5
PID    NAME          USER           CPU%  MEM%  STATE  COMMAND
5647   claude        user           50.2  8.6   S+     claude
612    iTerm2        user           40.8  4.1   S      /Applications/iTerm.app/Contents/MacOS/iTerm2
382    WindowServer  _windowserver  39.5  1.7   Ss     /System/Library/PrivateFrameworks/SkyLight.framework/...
21635  claude        user           27.0  14.3  S+     claude
71898  Magnet        user           14.1  0.7   S      /Applications/Magnet.app/Contents/MacOS/Magnet

$ pstop list --sort cpu
PID    NAME                  USER           CPU%  MEM%  STATE  COMMAND
382    WindowServer          _windowserver  37.7  1.7   Rs     /System/Library/PrivateFrameworks/...
612    iTerm2                user           22.5  4.1   R      /Applications/iTerm.app/...
72373  Claude                user           15.7  1.9   S      /Applications/Claude.app/...
87971  OrbStack              user           8.0   12.6  Ss     /Applications/OrbStack.app/...
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `list` | List all processes | `pstop list --sort cpu` |
| `top` | Top resource consumers | `pstop top --n 10` |
| `find <query>` | Find by name, command, or port | `pstop find node` |
| `info <pid>` | Detailed process info (files, ports, children) | `pstop info 1234` |
| `kill <pid>` | Kill process | `pstop kill 1234 --force` |
| `tree` | Process tree view | `pstop tree` |
| `dev` | Developer view grouped by stack | `pstop dev` |
| `watch <pid>` | Live-monitor a process | `pstop watch 1234 --interval 2` |

## TUI

Launch `pstop` without arguments for interactive mode:

- Live-updating process table (refreshes every 2s)
- Sort by CPU, MEM, PID, or Name (press `1`-`4`)
- Search/filter with `/`
- Tab switching: All | Top | Dev
- Kill selected process with `K` (with confirmation)
- View detailed info with `i`

Color coding: red for high CPU (>50%), yellow for medium (20-50%).

## Claude Code

Available as a skill in the [macos-toolkit](https://github.com/lu-zhengda/macos-toolkit) Claude Code plugin.

## License

[MIT](LICENSE)
