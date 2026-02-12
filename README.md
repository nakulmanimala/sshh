# SSHH

A fast, interactive terminal tool for managing and connecting to SSH servers. Built with Go and Bubble Tea.

## Features

- Interactive TUI with fuzzy search and filtering
- Add, edit, and delete server configurations
- Import hosts from `~/.ssh/config`
- Connection history with most-recently-used sorting
- Direct connect mode via CLI argument
- Clean SSH handoff using `syscall.Exec`

## Screenshots

<img width="490" height="283" alt="SCR-20260212-jrsg" src="https://github.com/user-attachments/assets/aaf091b0-75ed-45e3-9a9e-51b47b7aca1c" />

<br>
<br>

<img width="821" height="265" alt="SCR-20260212-jncn" src="https://github.com/user-attachments/assets/ccf54eb8-ba0b-4371-ac96-fe0f96bdb20e" />

<br>
<br>

<img width="617" height="169" alt="SCR-20260212-jpwj" src="https://github.com/user-attachments/assets/76b9cc96-28aa-47a6-b4ee-17d3a4faa95e" />

<br>

## Install

```bash
go install .
```

Or build manually:

```bash
go build -o sshh .
```

## Usage

Launch the interactive TUI:

```bash
./sshh
```

Connect directly to a saved server by name:

```bash
./sshh my-server
```

## Keybindings

| Key          | Action                    |
|--------------|---------------------------|
| `/`          | Search / filter servers   |
| `Enter`      | Connect to selected server|
| `a`          | Add a new server          |
| `e`          | Edit selected server      |
| `d`          | Delete selected server    |
| `i`          | Import from ~/.ssh/config |
| `q`          | Quit                      |

### Form (Add/Edit)

| Key              | Action              |
|------------------|----------------------|
| `Tab` / `Down`   | Next field           |
| `Shift+Tab` / `Up` | Previous field    |
| `Enter`          | Next field / save    |
| `Ctrl+S`         | Save                 |
| `Esc`            | Cancel               |

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

Connection history is tracked in `~/.sshh/history.json`.

## Requirements

- Go 1.21+
- SSH client installed and available in PATH
