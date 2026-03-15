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
		return fmt.Errorf("usage: dotfiles add <path> [path...]")
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

	for _, name := range names {
		if err := addOne(r, home, cwd, name); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", color.Red("error"), name, err)
		}
	}
	return nil
}

func addOne(r *repo.Repo, home, cwd, name string) error {
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
