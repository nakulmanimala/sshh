# SSHH

A fast, interactive terminal tool for managing and connecting to SSH servers and tunnels. Built with Go and Bubble Tea.

## Features

- Tabular TUI with live fuzzy search
- Manage SSH servers — add, edit, delete, import from `~/.ssh/config`
- Manage SSH tunnels — local, remote, and dynamic port forwarding
- Connection history with most-recently-used sorting
- Direct connect mode via CLI argument
- Clean SSH handoff using `syscall.Exec`

<br>
<br>

![ssh-demo](https://github.com/user-attachments/assets/c9546f32-6f45-4496-ad6d-b6c4310895ea)


<br>
<br>

![tunnel-demo](https://github.com/user-attachments/assets/f4364b1b-55cd-4331-aced-36adc9ade166)


<br>
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

Server configs are stored in `~/.sshh/config.yaml`:

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

Tunnel configs are stored in `~/.sshh/tunnels.yaml`:

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

Connection history is tracked in `~/.sshh/history.json`.

## Requirements

- Go 1.21+
- SSH client installed and available in PATH
