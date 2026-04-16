# dotfiles

Yet another dotfile manager — a small, opinionated CLI for managing dotfiles with symlinks and git.

There are dozens, maybe hundreds, of dotfile managers on GitHub. This is just my take on what I actually need: something lightweight, convention-based, and pragmatic. No YAML configs, no plugin systems, no dependency on a package manager. Just a single binary that moves files into a git repo, creates symlinks, and gets out of your way.

If you're happy with GNU Stow, chezmoi, or a bare git repo, those are great tools — keep using them. This exists because I wanted something simpler for my own workflow and figured others might too.

## How it works

The idea is simple:

1. You have a git repo (by default `~/.dotfiles`).
2. `dotfiles add` moves a file into that repo and leaves a symlink behind.
3. The repo structure mirrors your home directory, so `~/.zshrc` lives at `~/.dotfiles/.zshrc` and `~/.config/nvim/init.vim` lives at `~/.dotfiles/.config/nvim/init.vim`.
4. `dotfiles update` stages, commits, and pushes in one step.
5. On a new machine, `dotfiles clone` pulls the repo and restores all the symlinks automatically.

No config files. No mapping tables. The directory structure *is* the configuration.

## Quick start

Already have dotfiles you want to track? Here's zero to managed in about a minute:

```bash
# 1. Create the repo
mkdir ~/.dotfiles && cd ~/.dotfiles && git init

# 2. Add a remote (use your own repo)
git remote add origin git@github.com:you/dotfiles.git

# 3. Start tracking files
cd ~
dotfiles add .zshrc .vimrc .gitconfig

# 4. Push
dotfiles update -m "initial dotfiles"
```

That's it. Your originals are now symlinks pointing into the repo:

```
~/.zshrc -> ~/.dotfiles/.zshrc
~/.vimrc -> ~/.dotfiles/.vimrc
~/.gitconfig -> ~/.dotfiles/.gitconfig
```

## Install

### From source

Requires Go 1.23+.

```bash
git clone https://github.com/msjurset/dotfiles-cli.git
cd dotfiles-cli
make deploy
```

This builds the binary, copies it to `~/.local/bin/`, installs the man page, and sets up shell completions. Make sure `~/.local/bin` is in your `PATH`.

### Pre-built binaries

Download from the [releases page](https://github.com/msjurset/dotfiles-cli/releases) and put the binary somewhere in your `PATH`.

## Commands

### `dotfiles` (no arguments)

List dotfiles in the current directory, showing which ones are managed (symlinked into the repo) and which aren't.

```bash
$ cd ~
$ dotfiles
  managed   file  .zshrc     -> .dotfiles/.zshrc
  managed   file  .vimrc     -> .dotfiles/.vimrc
  managed   dir   .config    -> .dotfiles/.config
  unmanaged file  .bashrc
  unmanaged dir   .ssh
```

This is useful for seeing what's tracked at a glance without touching git.

### `dotfiles status`

A git-status-style view showing pending changes to managed files and any unmanaged candidates worth considering.

```bash
$ dotfiles status
Changes:
  M .zshrc
  M .config/nvim/init.vim

Unmanaged dotfiles:
  file  .bashrc
  dir   .ssh
```

If there are unpushed commits, it tells you that too:

```bash
$ dotfiles status
3 unpushed commit(s)

All managed dotfiles are clean.
```

### `dotfiles add <path> [path...]`

Move a file (or files) into the repo and replace the original with a relative symlink. Accepts multiple paths at once.

```bash
# Add a single file
$ dotfiles add .zshrc
  added .zshrc -> .dotfiles/.zshrc

# Add several at once
$ dotfiles add .vimrc .gitconfig .tmux.conf
  added .vimrc -> .dotfiles/.vimrc
  added .gitconfig -> .dotfiles/.gitconfig
  added .tmux.conf -> .dotfiles/.tmux.conf
```

Paths can be relative, absolute, or use `~/`:

```bash
dotfiles add .zshrc                       # relative to cwd
dotfiles add ~/.config/nvim/init.vim      # ~/ expansion
dotfiles add /Users/me/.bashrc            # absolute
```

Nested paths are preserved. Adding `~/.config/nvim/init.vim` creates `~/.dotfiles/.config/nvim/init.vim` and the necessary parent directories in the repo.

#### Adding files inside managed directories

If a directory is already managed (symlinked), new files created inside it are physically already in the repo. `dotfiles add` detects this and stages the file for git without trying to move or re-symlink anything:

```bash
# .zfuncs/ is already a symlink to .dotfiles/.zfuncs/
# You create a new function file:
$ cat > ~/.zfuncs/myfunc <<< 'function myfunc { echo hello }'

# dotfiles add detects it's already in the repo and just stages it
$ dotfiles add .zfuncs/myfunc
  tracked .zfuncs/myfunc (staged for commit)

# Now commit it
$ dotfiles update -m "add myfunc"
```

#### Safety

- **Collision detection**: if the repo already has a file at that path (and it's not the symlink case above), `add` refuses and leaves your original file untouched.
- **Rollback on failure**: if creating the symlink fails after the file has been moved, `add` moves it back.

### `dotfiles update [-m "message"]`

Stage all changes, commit, and push — the one-liner for "I edited some dotfiles, ship it."

```bash
# Auto-generated commit message based on changed filenames
$ dotfiles update
Changes:
  M .zshrc
Staged.
Committed: Update dotfiles: .zshrc
Pushed.

# Custom commit message
$ dotfiles update -m "add fzf key bindings"
Changes:
  M .zshrc
Staged.
Committed: add fzf key bindings
Pushed.
```

`update` is scope-aware:

- **Run from inside the repo** (`~/.dotfiles`): stages everything.
- **Run from any other directory** (e.g., `~`): only stages changes to dotfiles managed from that directory.

This means you can `cd ~` and run `dotfiles update` without worrying about accidentally committing work-in-progress changes to files managed from other directories.

### `dotfiles clone <repo> [path]`

Clone a dotfiles repo and automatically restore all symlinks. This is the "new machine bootstrap" command.

```bash
# Clone into the default location (~/.dotfiles)
$ dotfiles clone git@github.com:you/dotfiles.git
Cloning git@github.com:you/dotfiles.git into /Users/you/.dotfiles
Cloned.
  linked   .zshrc        -> ~/.zshrc
  linked   .vimrc        -> ~/.vimrc
  linked   .config/nvim/init.vim -> ~/.config/nvim/init.vim

3 linked, 0 already ok, 0 failed

# Clone into a custom location
$ dotfiles clone git@github.com:you/dotfiles.git ~/my-dotfiles
```

If a file already exists at the symlink target (e.g., the system created a default `.bashrc`), `clone` reports it as a conflict so you can decide what to do:

```bash
  conflict .bashrc (file exists, not a symlink)
```

### `dotfiles link [name...]`

Restore symlinks from repo contents. Useful after `clone` if something went wrong, or if you accidentally deleted a symlink.

```bash
# Restore all symlinks
$ dotfiles link
  ok       .zshrc
  ok       .vimrc
  linked   .tmux.conf -> ~/.tmux.conf

1 linked, 2 already ok, 0 failed

# Restore a specific symlink
$ dotfiles link .zshrc
  linked   .zshrc -> ~/.zshrc

1 linked, 0 already ok, 0 failed
```

`link` is idempotent — existing, correct symlinks are reported as "ok" and left alone. Stale symlinks (pointing to the wrong place) are replaced.

### `dotfiles clean [name...]`

Stop managing a dotfile. Removes the symlink and copies the file back from the repo to its original location. The file remains in the repo for git history purposes, but it's no longer symlinked.

```bash
# Unmanage a specific file
$ dotfiles clean .bashrc
  cleaned .bashrc (restored from repo)

1 dotfile(s) cleaned

# Unmanage everything in the current directory
$ cd ~
$ dotfiles clean
```

Like `add`, this has rollback protection — if copying the file back fails, the symlink is restored.

## Day-to-day workflow

Here's what typical daily usage looks like:

```bash
# Edit a managed dotfile — just edit it normally, the symlink is transparent
$ nvim ~/.zshrc

# Check what changed
$ dotfiles status
Changes:
  M .zshrc

# Ship it
$ dotfiles update
```

That's the whole loop: edit, status, update. Three commands, no ceremony.

### Adding a whole directory

Directories work the same as files:

```bash
$ dotfiles add .config/nvim
  added .config/nvim -> .dotfiles/.config/nvim
```

Now `~/.config/nvim` is a symlink. Any new files you create inside it (like adding a new plugin config) are automatically in the repo — just `dotfiles update` when you're ready to commit.

### Checking what's not tracked yet

```bash
$ cd ~
$ dotfiles status
Unmanaged dotfiles:
  file  .bashrc
  dir   .ssh
  file  .npmrc
```

The "unmanaged" section shows dotfiles in the current directory that aren't tracked yet. Not everything needs managing — `.ssh` probably shouldn't be in a public repo — but it's a handy reminder.

## New machine setup

The whole point of managing dotfiles is being able to set up a new machine quickly:

```bash
# 1. Install the binary (or build from source)
# 2. Clone your dotfiles
dotfiles clone git@github.com:you/dotfiles.git

# 3. Done. All your symlinks are restored.
```

If you have conflicts (files that already exist), resolve them manually and re-run `dotfiles link` for the ones you skipped.

## Shell completions

Tab completion is built in for both zsh and bash.

```bash
# Zsh (auto-installed by make deploy to oh-my-zsh custom completions)
dotfiles -completion zsh > ~/.oh-my-zsh/custom/completions/_dotfiles

# Bash
dotfiles -completion bash >> ~/.bashrc
```

Completions cover all subcommands, flags, and file arguments.

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DOTFILES_REPO_PATH` | `~/.dotfiles` | Path to the dotfiles git repository |

If your repo lives somewhere other than `~/.dotfiles`, set this in your shell config:

```bash
export DOTFILES_REPO_PATH="$HOME/my-dotfiles"
```

## Building from source

```bash
make build       # compile the binary
make test        # run all tests
make deploy      # build + install binary, man page, and completions
make release     # cross-compile for Linux, macOS, and Windows
```

## License

MIT
