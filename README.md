# dotfiles

A CLI tool for managing dotfiles with symlinks and git. Replaces GNU Stow with a simpler, purpose-built workflow.

## Features

- List dotfiles in the current directory with managed/unmanaged status
- Git-status-style view of pending changes and unmanaged candidates
- Move dotfiles into a git repo and replace them with relative symlinks
- Stage, commit, and push changes in a single command
- Clone and bootstrap dotfiles on a new machine in one command
- Restore symlinks from repo contents after cloning or accidental deletion
- Unmanage dotfiles by removing symlinks and restoring original files
- Scoped operations: works on cwd-relevant files when run outside the repo
- Convention-based: repo structure mirrors home-relative paths, no config files needed
- Configurable repo path via `DOTFILES_REPO_PATH` environment variable

## Install

```
make deploy
```

This builds the binary, installs it to `~/.local/bin/`, installs the man page, and sets up zsh completions.

## Usage

```
dotfiles                          # list dotfiles in cwd
dotfiles status                   # show changes and unmanaged candidates
dotfiles add <name> [name...]     # move dotfiles into repo, create symlinks
dotfiles update [-m "message"]    # git add, commit, push
dotfiles clone <repo> [path]      # clone repo and restore symlinks
dotfiles link [name...]           # restore symlinks from repo contents
dotfiles clean [name...]          # remove symlinks, restore files
```

### Commands

| Command | Description |
|---------|-------------|
| *(none)* | List dotfiles in cwd, showing managed vs unmanaged |
| `status` | Git-status-style view: changes to managed files + unmanaged candidates |
| `add <path>...` | Move file into repo, create relative symlink |
| `update [-m "msg"]` | Stage, commit, and push changes (scoped to cwd or all if in repo) |
| `clone <repo> [path]` | Clone a dotfiles repo (default `~/.dotfiles`) and restore all symlinks |
| `link [name...]` | Restore symlinks from repo contents (all or specific names) |
| `clean [name...]` | Remove symlinks and copy files back from repo |

### Options

| Option | Command | Description |
|--------|---------|-------------|
| `-m "message"` | `update` | Custom commit message. If omitted, auto-generates from changed filenames |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DOTFILES_REPO_PATH` | `~/.dotfiles` | Path to the dotfiles git repository |

### Examples

```bash
# See which dotfiles in ~ are managed
cd ~
dotfiles

# Check for uncommitted changes
dotfiles status

# Start managing your zsh and vim configs
dotfiles add .zshrc .vimrc

# Push changes after editing a managed dotfile
dotfiles update -m "update zsh aliases"

# Auto-generate commit message
dotfiles update

# Bootstrap on a new machine
dotfiles clone git@github.com:user/dotfiles.git

# Restore symlinks if one was deleted
dotfiles link .zshrc

# Restore all symlinks
dotfiles link

# Stop managing a dotfile (restores original file)
dotfiles clean .bashrc
```

## Build

```
make build
```

## License

MIT
