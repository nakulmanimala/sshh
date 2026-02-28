# SSHH
### The terminal UI your SSH setup deserves.

A fast, interactive terminal UI for managing and connecting to SSH servers and tunnels. Built with Go and Bubble Tea.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/nakulmanimala/sshh?style=for-the-badge&color=brightgreen)](https://github.com/nakulmanimala/sshh/releases)
[![License](https://img.shields.io/github/license/nakulmanimala/sshh?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey?style=for-the-badge&logo=apple)](https://github.com/nakulmanimala/sshh/releases)
[![Arch](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-blue?style=for-the-badge)](https://github.com/nakulmanimala/sshh/releases)

<br>

## Features

- Tabular TUI with live fuzzy search and real-time filtering
- Manage SSH servers — add, edit, delete, import from `~/.ssh/config`
- Manage SSH tunnels — local, remote, and dynamic port forwarding
- Connection history with most-recently-used sorting
- Customisable UI color theme per view, persisted across sessions
- Direct connect mode via CLI argument (`sshh my-server`)
- Clean SSH handoff using `syscall.Exec` — no wrapper process

<br>

![ssh-demo](https://github.com/user-attachments/assets/c9546f32-6f45-4496-ad6d-b6c4310895ea)

<br>
<br>

![tunnel-demo](https://github.com/user-attachments/assets/f4364b1b-55cd-4331-aced-36adc9ade166)

<br>

## Install

### Homebrew (recommended)

```bash
brew tap nakulmanimala/sshh
brew install sshh
```

### Go

```bash
go install github.com/nakulmanimala/sshh@latest
```

### Build from source

```bash
git clone https://github.com/nakulmanimala/sshh.git
cd sshh
go build -o sshh .
```

## Usage

Launch the TUI:

```bash
sshh
```

Connect directly to a saved server by name:

```bash
sshh my-server
```

Check version:

```bash
sshh --version
```

## Keybindings

### SSH Servers

| Key        | Action                      |
|------------|-----------------------------|
| `↑ / ↓`    | Navigate                    |
| `Enter`    | Connect to selected server  |
| `Tab`      | Switch to Tunnels view      |
| `Ctrl+V`   | Toggle list / table view    |
| `Ctrl+A`   | Add a new server            |
| `Ctrl+E`   | Edit selected server        |
| `Ctrl+D`   | Delete selected server      |
| `Ctrl+O`   | Import from ~/.ssh/config   |
| `Ctrl+T`   | Change theme color          |
| `Esc`      | Clear search                |
| `Ctrl+C`   | Quit                        |

### SSH Tunnels

| Key        | Action                      |
|------------|-----------------------------|
| `↑ / ↓`    | Navigate                    |
| `Enter`    | Run selected tunnel         |
| `Tab`      | Switch to Servers view      |
| `Ctrl+V`   | Toggle list / table view    |
| `Ctrl+A`   | Add a new tunnel            |
| `Ctrl+E`   | Edit selected tunnel        |
| `Ctrl+D`   | Delete selected tunnel      |
| `Ctrl+T`   | Change theme color          |
| `Esc`      | Clear search                |
| `Ctrl+C`   | Quit                        |

### Form (Add / Edit)

| Key                  | Action               |
|----------------------|----------------------|
| `Tab` / `Down`       | Next field           |
| `Shift+Tab` / `Up`   | Previous field       |
| `Ctrl+S`             | Save                 |
| `Esc`                | Cancel               |

## Configuration

All config files are stored in `~/.sshh/`.

| File | Purpose |
|------|---------|
| `~/.sshh/config.yaml` | SSH server list |
| `~/.sshh/tunnels.yaml` | SSH tunnel list |
| `~/.sshh/settings.yaml` | UI preferences (theme color) |
| `~/.sshh/history.json` | Connection history |

### SSH Servers (`~/.sshh/config.yaml`)

```yaml
servers:
  - name: my-server
    host: 192.168.1.10
    user: root
    port: 22
    key: ~/.ssh/id_rsa
    tags:
      - prod
      - web
```

### SSH Tunnels (`~/.sshh/tunnels.yaml`)

```yaml
tunnels:
  - name: db-tunnel
    ssh_host: 192.168.1.10
    ssh_user: root
    ssh_port: 22
    ssh_key: ~/.ssh/id_rsa
    type: local        # local | remote | dynamic
    local_port: 5432
    remote_host: localhost
    remote_port: 5432
```

## Requirements

- Go 1.25+ (for building from source)
- SSH client available in PATH

## License

MIT — see [LICENSE](LICENSE)
