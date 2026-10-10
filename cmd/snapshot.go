package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Snapshot(name string, list bool, deleteName string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	return RunSnapshot(r, name, list, deleteName)
}

func RunSnapshot(r *repo.Repo, name string, list bool, deleteName string) error {
	if list || name == "list" || name == "--list" || name == "-l" {
		snapshots, err := r.ListSnapshots()
		if err != nil {
			return err
		}
		if len(snapshots) == 0 {
			fmt.Println("No snapshots found.")
			fmt.Println("Create one with: dotfiles snapshot [name]")
			return nil
		}
		fmt.Println(color.Bold("Available snapshots:"))
		for _, s := range snapshots {
			fmt.Printf("  %s  %-20s %s\n",
				color.Green(fmt.Sprintf("%-24s", s.Name)),
				color.Dim(s.Date),
				color.Dim(s.Subject))
		}
		return nil
	}

	if deleteName != "" {
		if err := r.DeleteSnapshot(deleteName); err != nil {
			return err
		}
		fmt.Printf("%s Deleted snapshot %s.\n", color.Green("snapshot:"), color.Bold(deleteName))
		return nil
	}

	if name == "rm" || name == "delete" || name == "--delete" {
		return fmt.Errorf("missing snapshot name to delete (e.g. dotfiles snapshot --delete <name>)")
	}

	created, err := r.CreateSnapshot(name)
	if err != nil {
		return err
	}
	fmt.Printf("%s Created snapshot %s.\n", color.Green("snapshot:"), color.Bold(created))
	return nil
}

func Rollback(name string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return RunRollback(r, home, name)
}

func RunRollback(r *repo.Repo, home, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("missing snapshot name (usage: dotfiles rollback <name>)\nRun 'dotfiles snapshot --list' to see available snapshots")
	}

	name = strings.TrimPrefix(name, "snapshot/")
	tagName := "snapshot/" + name
	if !r.TagExists(tagName) {
		return fmt.Errorf("snapshot %q not found (use 'dotfiles snapshot --list' to see available snapshots)", name)
	}

	// 1. Create safety snapshot before rolling back
	safetyName := fmt.Sprintf("pre-rollback-%s", time.Now().Format("20060102-150405"))
	_, _ = r.CreateSnapshot(safetyName)

	// 2. Remove files added between snapshot and HEAD
	diffOut, err := r.Git("diff", "--name-only", "--diff-filter=A", tagName, "HEAD")
	if err == nil && strings.TrimSpace(diffOut) != "" {
		for _, rel := range strings.Split(diffOut, "\n") {
			rel = strings.TrimSpace(rel)
			if rel != "" {
				_ = os.Remove(filepath.Join(r.Path, rel))
			}
		}
	}

	// 3. Checkout snapshot files
	_, err = r.Git("checkout", tagName, "--", ".")
	if err != nil {
		return fmt.Errorf("restoring files from snapshot %s: %w", name, err)
	}

	// 4. Update symlinks in home
	entries, err := walkRepo(r)
	if err == nil && len(entries) > 0 {
		_ = linkEntriesToHome(r, home, entries, false)
	}

	// 5. Remove any broken symlinks in home that pointed into repo
	homeEntries, err := dotfile.Scan(home, r.Path)
	if err == nil {
		for _, ent := range homeEntries {
			if ent.Managed {
				if _, err := os.Stat(ent.AbsPath); os.IsNotExist(err) {
					_ = os.Remove(ent.AbsPath)
					fmt.Printf("  %s removed dangling symlink %s\n", color.Yellow("unlinked"), ent.Name)
				}
			}
		}
	}

	fmt.Printf("%s Successfully rolled back to snapshot %s.\n", color.Green("rollback:"), color.Bold(name))
	fmt.Printf("Safety snapshot created: %s (run 'dotfiles rollback %s' to undo).\n",
		color.Cyan(safetyName), safetyName)
	return nil
}
