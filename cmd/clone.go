package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

func Clone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dotfiles clone <repository> [path]")
	}
	repository := args[0]

	// Determine destination path
	var dest string
	if len(args) >= 2 {
		dest = args[1]
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}
		dest = filepath.Join(home, ".dotfiles")
	}

	// Expand ~ if present
	if strings.HasPrefix(dest, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home dir: %w", err)
		}
		dest = filepath.Join(home, dest[2:])
	}

	// Check destination doesn't already exist
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%s already exists", dest)
	}

	fmt.Printf("Cloning %s into %s\n", color.Cyan(repository), color.Bold(dest))
	cmd := exec.Command("git", "clone", repository, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	fmt.Println(color.Green("Cloned."))

	// Restore symlinks by walking the repo
	r := &repo.Repo{Path: dest}
	entries, err := walkRepo(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s could not walk repo: %v\n", color.Yellow("warning:"), err)
		fmt.Println("Run 'dotfiles link' to restore symlinks manually.")
		return nil
	}
	if len(entries) == 0 {
		fmt.Println("No files found in repo. Use 'dotfiles add' to start tracking dotfiles.")
		return nil
	}

	return linkEntries(r, entries, false)
}
