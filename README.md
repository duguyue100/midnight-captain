# Midnight Captain

A dual-pane terminal file manager with Vim keybindings and a TokyoNight theme. Inspired by Midnight Commander. Single binary, no config file required.

---

## Features

- **Dual pane** — navigate two directories side by side, copy/move between them
- **Vim keybindings** — `j`/`k`/`h`/`l`, `gg`/`G`, `ctrl+d`/`ctrl+u`
- **Visual selection** — select ranges with `V`, toggle individual files with `Space`
- **File operations** — copy, cut, paste, delete (with confirmation), rename
- **Fuzzy search** — `/` opens a recursive file search in the current directory
- **Command palette** — `:` opens a command prompt for power-user actions
- **SSH support** — browse and operate on remote servers over SFTP via `/ssh user@host`
- **Nerd font icons** — file type icons for 40+ formats
- **TokyoNight theme** — easy on the eyes

---

## Requirements

- A terminal with [Nerd Fonts](https://www.nerdfonts.com/) support for icons (recommended: JetBrainsMono Nerd Font, FiraCode Nerd Font)
- macOS or Linux

---

## Installation

### Download binary

Download the latest binary from the [releases page](https://github.com/dgyhome/midnight-captain/releases) and place it on your `$PATH`:

```sh
curl -Lo mc https://github.com/dgyhome/midnight-captain/releases/latest/download/mc-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
chmod +x mc
mv mc ~/.local/bin/mc
```

### Build from source

Requires [Go 1.22+](https://go.dev/dl/).

```sh
git clone https://github.com/dgyhome/midnight-captain
cd midnight-captain
make install        # builds and copies mc to ~/.local/bin/mc
```

Or build without installing:

```sh
make build          # binary lands at bin/mc
./bin/mc
```

Ensure `~/.local/bin` is on your `$PATH`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

---

## Usage

```sh
mc
```

Launches in the current directory with dual panes. On terminals narrower than 80 columns, a single pane is shown instead.

---

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `h` / `←` / `Backspace` | Go to parent directory |
| `l` / `→` / `Enter` | Enter directory or open file |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+d` | Half page down |
| `Ctrl+u` | Half page up |
| `Tab` | Switch active pane |
| `.` | Toggle hidden files |

### Selection

| Key | Action |
|-----|--------|
| `V` | Enter visual mode — move to extend selection |
| `Space` | Toggle selection on current item |
| `Esc` | Clear selection / exit visual mode |

**Range select:** press `V` on the first file, then navigate with `j`/`k`/`G`/`gg` — all files between the anchor and cursor are selected automatically.

### File Operations

| Key | Action |
|-----|--------|
| `y` | Yank (copy) selected files to clipboard |
| `d` | Cut selected files to clipboard |
| `p` | Paste clipboard into active pane |
| `x` | Delete selected files (prompts for confirmation) |
| `r` | Rename current file |
| `e` | Open current file in `nvim` |

### Search & Commands

| Key | Action |
|-----|--------|
| `/` | Fuzzy search files recursively from current directory |
| `:` | Open command palette |
| `q` | Quit |

### Search Overlay

| Key | Action |
|-----|--------|
| _(type)_ | Filter results |
| `Ctrl+j` / `↓` | Next result |
| `Ctrl+k` / `↑` | Previous result |
| `Enter` | Navigate to selected file |
| `Esc` | Close |

### Command Palette

| Key | Action |
|-----|--------|
| _(type)_ | Filter commands |
| `Ctrl+j` / `↓` | Next suggestion |
| `Ctrl+k` / `↑` | Previous suggestion |
| `Tab` | Autocomplete |
| `Enter` | Execute |
| `Esc` | Close |

---

## Commands

Open with `:`, then type a command name.

| Command | Description |
|---------|-------------|
| `/ssh user@host` | Connect to a remote server (SFTP) in the active pane |
| `/disconnect` | Disconnect SSH, return to local filesystem |
| `/sort name\|size\|date` | Change sort order |
| `/hidden` | Toggle hidden files |
| `/mkdir [name]` | Create a new directory |
| `/touch [name]` | Create an empty file |
| `/quit` | Exit |

---

## SSH

Connect to a remote server:

```
:  /ssh user@192.168.1.10
```

Authentication is tried in order: SSH agent → key files (`~/.ssh/id_ed25519`, `~/.ssh/id_rsa`) → known hosts verification. The active pane switches to the remote filesystem. All file operations (copy, move, delete, rename) work across local and remote panes.

Disconnect:

```
:  /disconnect
```

---

## License

MIT
