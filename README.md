<img width="2320" height="464" alt="Gemini_Generated_Image_618bk5618bk5618b" src="https://github.com/user-attachments/assets/b28dc1cb-04e1-43c7-b3f8-41f493e10aaa" />

---

## Features

- **Dual pane** — navigate two directories side by side, copy/move between them
- **Tree expand** — expand/collapse directories inline with `l`/`h`, nerd font icons
- **Vim keybindings** — `j`/`k`/`h`/`l`, `gg`/`G`, `ctrl+d`/`ctrl+u`
- **Visual selection** — select ranges with `V`, operate on multiple files at once
- **File operations** — copy, cut, paste, delete (with confirmation), rename
- **Smart create** — `a` creates a file or directory (trailing `/` = dir, nested paths supported)
- **Fuzzy search** — `space` opens fuzzy search in current dir; `:find` searches recursively
- **Command palette** — `:` opens a command prompt for power-user actions
- **Goto** — `:goto <path>` with live directory listing, tab-complete, and `~` expansion
- **SSH** — browse and operate on remote servers over SFTP via `:ssh user@host`
- **Nerd font icons** — file type icons, open/closed folder glyphs
- **TokyoNight theme** — easy on the eyes

## Screenshot

<img width="1709" height="1042" alt="image" src="https://github.com/user-attachments/assets/148a50d7-6b43-4a41-9dc1-99f1c63504e6" />

---

## Requirements

- A terminal with [Nerd Fonts](https://www.nerdfonts.com/) support (recommended: JetBrainsMono Nerd Font, FiraCode Nerd Font)
- macOS or Linux

---

## Installation

### One-liner (macOS and Linux)

```sh
curl -fsSL https://raw.githubusercontent.com/duguyue100/midnight-captain/main/install.sh | bash
```

Downloads the correct pre-built binary for your OS and architecture from the latest GitHub Release and places it at `~/.local/bin/mc`. If `~/.local/bin` is not in your `$PATH`, the installer will tell you what to add.

### Build from source (local)

Requires [Go 1.22+](https://go.dev/dl/).

```sh
git clone https://github.com/duguyue100/midnight-captain
cd midnight-captain
./install.sh --local-build
```

Runs `make build` and copies `bin/mc` to `~/.local/bin/mc`.

### Manual build

```sh
make build   # binary at bin/mc
./bin/mc
```

---

## Usage

```sh
mc
```

Launches in the current directory with dual panes.

---

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `ctrl+d` / `ctrl+u` | Half-page down / up |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `tab` | Switch active pane |
| `h` | Collapse directory or jump to parent |
| `l` / `enter` | Expand directory inline |
| `o` | Navigate into directory (change cwd) |
| `.` | Toggle hidden files |

### Selection

| Key | Action |
|-----|--------|
| `V` | Visual select mode — move to extend selection |
| `esc` | Cancel / clear selection |

### File Operations

| Key | Action |
|-----|--------|
| `a` | Smart create — trailing `/` = dir, nested paths supported |
| `r` | Rename current file |
| `y` | Yank (copy) to clipboard |
| `d` | Cut to clipboard |
| `p` | Paste into active pane |
| `x` | Delete (prompts for confirmation) |
| `e` | Open in `nvim` |

### Search & Commands

| Key | Action |
|-----|--------|
| `space` | Fuzzy search in current directory |
| `:` | Open command palette |
| `?` | Show help overlay |
| `q` | Quit |

---

## Commands

Open with `:`, then type a command.

| Command | Description |
|---------|-------------|
| `:ssh user@host` | Connect to remote server (SFTP) in active pane |
| `:disconnect` | Disconnect SSH, return to local filesystem |
| `:goto <path>` | Jump to path with live completion |
| `:find` | Recursive fuzzy search from current directory |
| `:sort name\|size\|date` | Change sort order |
| `:hidden` | Toggle hidden files |
| `:quit` | Exit |

---

## SSH

Connect to a remote server:

```
:ssh user@192.168.1.10
```

Auth order: SSH agent → key files (`~/.ssh/id_ed25519`, `~/.ssh/id_rsa`) → password prompt.
The active pane switches to the remote filesystem. File operations work across local↔remote panes.

Disconnect:

```
:disconnect
```

---

## License

MIT — see [LICENSE](LICENSE).
