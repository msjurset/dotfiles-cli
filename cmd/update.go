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

func Update(message string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	inRepo, _ := isInside(r.Path, cwd)

	var paths []string
	if inRepo {
		paths = []string{"."}
	} else {
		entries, err := dotfile.Scan(cwd, r.Path)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if e.Managed {
				paths = append(paths, e.RepoRel)
			}
		}
	}

	if len(paths) == 0 {
		fmt.Println("No managed dotfiles found.")
		return nil
	}

	// Check for changes
	out, err := r.StatusPorcelain(paths...)
	if err != nil {
		return err
	}
	if out == "" {
		fmt.Println(color.Green("Nothing to update."))
		return nil
	}

	fmt.Println(color.Bold("Changes:"))
	for _, line := range strings.Split(out, "\n") {
		if line != "" {
			fmt.Printf("  %s\n", colorStatus(line))
		}
	}

	// Stage
	if err := r.Add(paths...); err != nil {
		return fmt.Errorf("staging: %w", err)
	}
	fmt.Println(color.Green("Staged."))

	// Build commit message
	if message == "" {
		var files []string
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Porcelain format: XY filename
			if len(line) > 3 {
				files = append(files, filepath.Base(strings.TrimSpace(line[2:])))
			}
		}
		message = "Update dotfiles: " + strings.Join(files, ", ")
	}

	if err := r.Commit(message); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	fmt.Printf("%s %s\n", color.Green("Committed:"), message)

	if err := r.PullRebase(); err != nil {
		return fmt.Errorf("pull --rebase: %w", err)
	}

	if err := r.Push(); err != nil {
		return fmt.Errorf("push: %w", err)
	}
	fmt.Println(color.Green("Pushed."))
	return nil
}
