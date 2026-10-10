package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/msjurset/dotfiles/cmd"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) == 0 {
		return cmd.List(false)
	}

	for _, a := range args {
		if a == "-h" || a == "-help" || a == "--help" {
			printUsage()
			return nil
		}
	}

	switch args[0] {
	case "-version", "--version":
		fmt.Printf("dotfiles %s\n", version)
		return nil
	case "-completion", "--completion":
		if len(args) < 2 {
			return fmt.Errorf("usage: dotfiles -completion <zsh|bash>")
		}
		switch args[1] {
		case "zsh":
			fmt.Print(zshCompletion)
		case "bash":
			fmt.Print(bashCompletion)
		default:
			return fmt.Errorf("unsupported shell: %s (supported: zsh, bash)", args[1])
		}
		return nil
	case "-h", "-help", "--help":
		printUsage()
		return nil
	case "-a", "--all":
		return cmd.List(true)
	case "status":
		showAll := false
		for _, a := range args[1:] {
			if a == "-a" || a == "--all" {
				showAll = true
			}
		}
		return cmd.Status(showAll)
	case "list":
		showAll := false
		for _, a := range args[1:] {
			if a == "-a" || a == "--all" {
				showAll = true
			}
		}
		return cmd.List(showAll)
	case "diff":
		return cmd.Diff(args[1:])
	case "add":
		return cmd.Add(args[1:])
	case "sync":
		msg := ""
		force := false
		for i := 1; i < len(args); i++ {
			if (args[i] == "-m" || args[i] == "--message") && i+1 < len(args) {
				msg = args[i+1]
				i++
			} else if args[i] == "-f" || args[i] == "--force" {
				force = true
			}
		}
		return cmd.Sync(msg, force)
	case "update":
		msg := ""
		force := false
		for i := 1; i < len(args); i++ {
			if (args[i] == "-m" || args[i] == "--message") && i+1 < len(args) {
				msg = args[i+1]
				i++
			} else if args[i] == "-f" || args[i] == "--force" {
				force = true
			}
		}
		return cmd.Update(msg, force)
	case "doctor":
		fix := false
		for _, a := range args[1:] {
			if a == "--fix" || a == "-f" {
				fix = true
			}
		}
		return cmd.Doctor(fix)
	case "audit":
		return cmd.Audit(args[1:])
	case "which":
		if len(args) < 2 {
			return fmt.Errorf("usage: dotfiles which <name>")
		}
		return cmd.Which(args[1])
	case "deps":
		strict := false
		for _, a := range args[1:] {
			if a == "--strict" || a == "-s" {
				strict = true
			}
		}
		return cmd.Deps(strict)
	case "profile":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		return cmd.Profile(name)
	case "browse", "fzf":
		printOnly := false
		query := ""
		for _, a := range args[1:] {
			if a == "--print" || a == "-p" {
				printOnly = true
			} else if query == "" && !strings.HasPrefix(a, "-") {
				query = a
			}
		}
		return cmd.Browse(query, printOnly)
	case "snapshot":
		list := false
		del := ""
		name := ""
		for i := 1; i < len(args); i++ {
			if args[i] == "--list" || args[i] == "-l" || args[i] == "list" {
				list = true
			} else if (args[i] == "--delete" || args[i] == "-d" || args[i] == "rm" || args[i] == "delete") && i+1 < len(args) {
				del = args[i+1]
				i++
			} else if name == "" && !strings.HasPrefix(args[i], "-") {
				name = args[i]
			}
		}
		return cmd.Snapshot(name, list, del)
	case "rollback":
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		return cmd.Rollback(name)
	case "edit":
		query := ""
		if len(args) > 1 {
			query = args[1]
		}
		return cmd.Edit(query)
	case "log":
		return cmd.Log(args[1:])
	case "ignore":
		return cmd.Ignore(args[1:])
	case "guard":
		return cmd.Guard(args[1:])
	case "clone":
		return cmd.Clone(args[1:])
	case "link":
		return cmd.Link(args[1:])
	case "clean":
		return cmd.Clean(args[1:])
	default:
		return fmt.Errorf("unknown command: %s\nUsage: dotfiles [status|diff|add|sync|update|doctor|audit|which|deps|profile|browse|snapshot|rollback|edit|log|ignore|guard|clone|link|clean]", args[0])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `dotfiles %s

Usage: dotfiles [command] [options]

Commands:
  (none)          List dotfiles in current directory
  status          Show changes and unmanaged candidates (-a for all)
  diff            Show changes against git repo (--staged for cached)
  add <path>      Move file into repo, create symlink (-f for sensitive)
  sync            Pull remote changes, restore symlinks, commit & push
  update          Stage, commit, and push changes (-f to override secrets)
  doctor          Check health of symlinks and permissions (--fix to repair)
  audit           Scan managed dotfiles for sensitive credentials/keys
  which <name>    Trace where an alias, function, or export is defined
  deps            Audit external command dependencies across shell configs
  profile [name]  Manage machine configuration overlays (--clear to deactivate)
  browse [query]  Interactive fuzzy finder with live preview and diffs (fzf)
  snapshot [name] Create or list tagged config checkpoints (--list, -d)
  rollback <name> Revert repo and symlinks to a saved snapshot
  edit [query]    Open matching dotfile directly in $EDITOR
  log [path]      Show commit history for dotfiles repo or specific file
  ignore          View or add patterns to .dotfilesignore
  guard           View or add sensitive guardrail patterns to .dotfilesguard
  clone <repo>    Clone dotfiles repo and restore symlinks
  link [name]     Restore symlinks from repo contents
  clean [name]    Remove symlinks, restore original files

Options:
  -a, --all       Show all files including ignored
  -f, --force     Override sensitive guardrails / allow secrets
  -m "msg"        Commit message (for update and sync commands)
  --fix           Automatically repair issues found by doctor
  --strict, -s    Exit with non-zero status if missing dependencies (deps)
  --list, -l      List saved snapshots (snapshot)
  --delete, -d    Delete specified snapshot (snapshot)
  --print, -p     Print selected file path instead of opening in $EDITOR (browse)
  -n <count>      Number of commits to display (for log command)
  -version        Print version and exit
  -completion     Print shell completion (zsh, bash)
  -help           Show this help
`, version)
}
