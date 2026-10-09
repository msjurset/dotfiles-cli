package cmd

import (
	"fmt"
	"os"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Sync(message string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	fmt.Println("Pulling latest changes from remote...")
	if err := r.PullRebase(); err != nil {
		return fmt.Errorf("pull --rebase: %w", err)
	}
	fmt.Println(color.Green("Repository up to date with remote."))

	// Automatically link any new or missing dotfiles
	fmt.Println("Checking symlinks...")
	if err := Link(nil); err != nil {
		return fmt.Errorf("restoring symlinks: %w", err)
	}

	// Check if there are local working tree changes to commit and push
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

	out, err := r.StatusPorcelain(paths...)
	if err != nil {
		return err
	}

	if out != "" {
		fmt.Println("\nLocal changes detected, committing and pushing...")
		return Update(message)
	}

	// Push any existing unpushed commits
	unpushed, err := r.UnpushedCount()
	if err == nil && unpushed > 0 {
		fmt.Printf("Pushing %d local commit(s)...\n", unpushed)
		if err := r.Push(); err != nil {
			return fmt.Errorf("push: %w", err)
		}
		fmt.Println(color.Green("Pushed."))
	}

	fmt.Println(color.Green("All dotfiles in sync!"))
	return nil
}
