package main

import (
	"fmt"
	"os"

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
		return fmt.Errorf("unknown command: %s\nUsage: dotfiles [status|diff|add|sync|update|doctor|audit|edit|log|ignore|guard|clone|link|clean]", args[0])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `dotfiles %s

Usage: dotfiles [command] [options]

Commands:
  (none)       List dotfiles in current directory
  status       Show changes and unmanaged candidates (-a for all)
  diff         Show changes against git repo (--staged for cached)
  add <path>   Move file into repo, create symlink (-f for sensitive)
  sync         Pull remote changes, restore symlinks, commit & push
  update       Stage, commit, and push changes (-f to override secrets)
  doctor       Check health of symlinks and permissions (--fix to repair)
  audit        Scan managed dotfiles for sensitive credentials/keys
  edit [query] Open matching dotfile directly in $EDITOR
  log [path]   Show commit history for dotfiles repo or specific file
  ignore       View or add patterns to .dotfilesignore
  guard        View or add sensitive guardrail patterns to .dotfilesguard
  clone <repo> Clone dotfiles repo and restore symlinks
  link [name]  Restore symlinks from repo contents
  clean [name] Remove symlinks, restore original files

Options:
  -a, --all    Show all files including ignored
  -f, --force  Override sensitive guardrails / allow secrets
  -m "msg"     Commit message (for update and sync commands)
  --fix        Automatically repair issues found by doctor
  -n <count>   Number of commits to display (for log command)
  -version     Print version and exit
  -completion  Print shell completion (zsh, bash)
  -help        Show this help
`, version)
}
