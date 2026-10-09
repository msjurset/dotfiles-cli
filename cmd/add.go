package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Add(names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("usage: dotfiles add [--force|-f] <path> [path...]")
	}
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	force := false
	var paths []string
	for _, n := range names {
		if n == "--force" || n == "-f" {
			force = true
		} else {
			paths = append(paths, n)
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("usage: dotfiles add [--force|-f] <path> [path...]")
	}

	for _, name := range paths {
		if err := addOne(r, home, cwd, name, force); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", color.Red("error"), name, err)
		}
	}
	return nil
}

func addOne(r *repo.Repo, home, cwd, name string, force bool) error {
	// Resolve to absolute path: support absolute paths, ~/paths, and relative paths
	absPath := name
	if strings.HasPrefix(absPath, "~/") {
		absPath = filepath.Join(home, absPath[2:])
	} else if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(cwd, absPath)
	}
	absPath = filepath.Clean(absPath)

	// Validate existence
	info, err := os.Lstat(absPath)
	if err != nil {
		return fmt.Errorf("%s does not exist", name)
	}

	// Must be under $HOME
	rel, err := filepath.Rel(home, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("%s is not under home directory", name)
	}

	// Sensitive guardrail check
	guardPatterns, _ := dotfile.LoadGuardPatterns(r.Path)
	if sensitive, pattern := dotfile.IsSensitive(absPath, guardPatterns); sensitive {
		if !force {
			return fmt.Errorf("matches sensitive guardrail pattern %q (use --force or -f to override)", pattern)
		}
		fmt.Printf("  %s %s: matched sensitive guardrail %q (forced)\n",
			color.Yellow("warning"), color.Bold(name), pattern)
	}

	// If the file's real path (resolving parent symlinks) is already inside the
	// repo, it doesn't need moving or symlinking — just stage it for git.
	// This handles files created inside a symlinked directory (e.g. .zfuncs/).
	if inside, repoRel := isInside(r.Path, absPath); inside {
		if err := r.Add(repoRel); err != nil {
			return fmt.Errorf("staging %s: %w", repoRel, err)
		}
		fmt.Printf("  %s %s %s\n",
			color.Green("tracked"),
			color.Bold(name),
			color.Dim("(staged for commit)"))
		return nil
	}

	// Check not already managed
	if info.Mode()&os.ModeSymlink != 0 {
		managed, _ := dotfile.IsManaged(absPath, r.Path)
		if managed {
			return fmt.Errorf("%s is already managed", name)
		}
	}

	// Detect collision: if the repo destination already exists
	repoRel := rel
	repoDest := filepath.Join(r.Path, repoRel)
	if _, err := os.Stat(repoDest); err == nil {
		return fmt.Errorf("collision: %s already exists in repo at %s", filepath.Base(name), repoRel)
	}

	// Create parent dirs in repo
	if err := os.MkdirAll(filepath.Dir(repoDest), 0755); err != nil {
		return fmt.Errorf("creating repo dirs: %w", err)
	}

	// Move into repo
	if err := os.Rename(absPath, repoDest); err != nil {
		return fmt.Errorf("moving %s to repo: %w", name, err)
	}

	// Create relative symlink
	if err := dotfile.CreateRelativeSymlink(absPath, repoDest); err != nil {
		// Rollback: move back
		_ = os.Rename(repoDest, absPath)
		return fmt.Errorf("creating symlink: %w", err)
	}

	fmt.Printf("  %s %s %s %s\n",
		color.Green("added"),
		color.Bold(name),
		color.Dim("->"),
		color.Cyan(".dotfiles/"+repoRel))
	return nil
}
