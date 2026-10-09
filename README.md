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

### `dotfiles [-a|--all]` (no arguments or `list`)

List dotfiles in the current directory, showing which ones are managed (symlinked into the repo) and which aren't. Common cache, runtime, and OS files (`.DS_Store`, `.Trash`, `.*_history`, `.cache`, `.npm`, etc.) are hidden by default to keep the output clean. Pass `-a` or `--all` to show all files including ignored items.

```bash
$ cd ~
$ dotfiles
  managed   file  .zshrc     -> .dotfiles/.zshrc
  managed   file  .vimrc     -> .dotfiles/.vimrc
  managed   dir   .config    -> .dotfiles/.config
  unmanaged file  .bashrc

(32 unmanaged dotfile(s) hidden by ignore rules; use -a or --all to show)
```

When run inside the dotfiles repository (`~/.dotfiles`), it automatically recognizes the repo environment and displays tracked files instead of listing repo files as unmanaged.

### `dotfiles status [-a|--all]`

A git-status-style view showing pending changes to managed files and unmanaged candidates worth considering. Filtered by ignore patterns (use `-a` or `--all` to view all).

```bash
$ dotfiles status
Changes:
  M .zshrc
  M .config/nvim/init.vim

Unmanaged dotfiles:
  file  .bashrc

(32 unmanaged dotfile(s) hidden by ignore rules; use -a or --all to show)
```

If there are unpushed commits, it tells you that too:

```bash
$ dotfiles status
3 unpushed commit(s)

All managed dotfiles are clean.
```

When run inside `~/.dotfiles`, `status` reports the full git working tree status and unpushed commits directly.

### `dotfiles diff [--staged|--cached] [name...]`

View changes made to managed dotfiles against the git repository. Accepts optional dotfile names and supports `--staged` (or `--cached`) to inspect staged modifications.

```bash
# View working tree diff for all modified managed files
$ dotfiles diff

# View diff for a specific file
$ dotfiles diff .zshrc

# View staged changes
$ dotfiles diff --staged
```

### `dotfiles add [--force|-f] <path> [path...]`

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

#### Sensitive File Guardrails

To prevent accidental leaks of credentials and secrets, `dotfiles add` automatically blocks files matching sensitive patterns (`.ssh`, `.gnupg`, `id_rsa*`, `id_ed25519*`, `*.pem`, `*.key`, `.env*`, `.secrets*`, `.aws`, etc.):

```bash
$ dotfiles add .env.production
  error .env.production: matches sensitive guardrail pattern ".env*" (use --force or -f to override)
```

Pass `--force` or `-f` to explicitly override the guardrail when tracking is genuinely intended:

```bash
$ dotfiles add -f .env.template
  warning .env.template: matched sensitive guardrail ".env*" (forced)
  added .env.template -> .dotfiles/.env.template
```

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

### `dotfiles sync [-m "message"]`

Pull remote updates (`git pull --rebase`), automatically restore any missing or new symlinks (`dotfiles link`), and commit/push local modifications in a single command. Ideal for moving between multiple workstations.

```bash
$ dotfiles sync
Pulling latest changes from remote...
Repository up to date with remote.
Checking symlinks...
  linked   .alacritty.toml -> ~/.alacritty.toml
1 linked, 12 already ok, 0 failed
All dotfiles in sync!
```

### `dotfiles update [-m "message"]`

Stage all changes, commit, pull any upstream remote changes with rebase, and push — the one-liner for "I edited some dotfiles, ship it."

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

### `dotfiles doctor [--fix]`

Audit your dotfiles setup for drift, clobbered symlinks, dangling pointers, and insecure permissions:

```bash
$ dotfiles doctor
Checking dotfiles health...
  [CLOBBERED] .zshrc in ~ is a regular file, not a symlink to repo
  [MISSING]   .tmux.conf is missing symlink in home directory

doctor: detected 2 issue(s). Run 'dotfiles doctor --fix' to automatically repair them.
```

Passing `--fix` (or `-f`) automatically repairs detected issues:
- **Clobbered symlinks:** Safely backs up modified local files (`.backup.<timestamp>`) and restores the symlink.
- **Missing symlinks:** Restores missing symlinks and creates parent directories as needed.
- **Misdirected symlinks:** Re-points incorrect symlinks to the proper repository target.
- **Permissions:** Tightens overly permissive `~/.ssh` or `~/.gnupg` directories.

```bash
$ dotfiles doctor --fix
Applying fixes...
    Backed up existing file to .zshrc.backup.1791587408
  fixed ~/.zshrc
  fixed ~/.tmux.conf
Complete: Repaired 2 of 2 issue(s).
```

### `dotfiles audit [path...]`

Scan your dotfiles for accidentally committed secrets and credentials. Inspects files for private key headers, AWS access keys, GitHub personal access tokens, Google API keys, OpenAI keys, and Slack tokens:

```bash
$ dotfiles audit
WARNING .env.backup:
  line 4: AWS Access Key ID (AKIA...MPLE)
  line 8: GitHub Personal Access Token (ghp_...bcde)

error: audit detected 2 potential credential(s)
```

The secret scanner also runs automatically during `dotfiles update` and `dotfiles sync` to prevent committing or pushing credentials to remote repositories. Use `-f` or `--force` if you intentionally need to commit an exempted pattern.

### `dotfiles edit [query]`

Quickly jump straight into editing any managed dotfile in `$EDITOR` (or `$VISUAL`, `nvim`, `vim`, `nano`) regardless of your current directory:

```bash
dotfiles edit zsh        # matches ~/.dotfiles/.zshrc
dotfiles edit alias      # matches ~/.dotfiles/.aliases
dotfiles edit nvim       # matches ~/.dotfiles/.config/nvim/init.vim
dotfiles edit tmux       # matches ~/.dotfiles/.tmux.conf
```

If run with no arguments, `dotfiles edit` opens the root repository directory in your editor. Upon exiting, if the file was modified, `edit` reminds you to run `dotfiles update` to commit and push.

### `dotfiles log [-n <count>] [path]`

View a compact, formatted history of your configuration changes with relative timestamps and author info:

```bash
$ dotfiles log -n 5
  9ea22e8  (3 hours ago)   Mark Sjurseth   Add in-repo awareness, ignore filtering, guardrails
  973780d  (yesterday)     Mark Sjurseth   Add caf alias and mybrewapps speedup
  dd825b8  (2 days ago)    Mark Sjurseth   Add stow compatibility docs

# View history specific to a file
$ dotfiles log .zshrc
```

### `dotfiles ignore [pattern...]`

View active ignore patterns or append new ignore rules to `~/.dotfiles/.dotfilesignore`. Automatically stages the ignore file for git.

```bash
# View active ignore rules (defaults + custom)
$ dotfiles ignore

# Add new ignore patterns
$ dotfiles ignore "*.tmp" ".mypy_cache"
  ignored *.tmp (added to .dotfilesignore)
  ignored .mypy_cache (added to .dotfilesignore)
```

### `dotfiles guard [pattern...]`

View sensitive guardrails or append custom sensitive patterns to `~/.dotfiles/.dotfilesguard`. Automatically stages the guard file for git.

```bash
# View active guardrails
$ dotfiles guard

# Guard additional sensitive files
$ dotfiles guard "*.pfx" "server.key"
  guarded *.pfx (added to .dotfilesguard)
  guarded server.key (added to .dotfilesguard)
```

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
