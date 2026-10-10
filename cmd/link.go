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

func Link(names []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	entries, err := walkRepo(r)
	if err != nil {
		return fmt.Errorf("walking repo: %w", err)
	}
	if len(entries) == 0 {
		fmt.Println("No files found in repo. Use 'dotfiles add' to start tracking dotfiles.")
		return nil
	}

	// If specific names given, filter
	if len(names) > 0 {
		var filtered []string
		for _, name := range names {
			found := false
			for _, repoRel := range entries {
				if repoRel == name || filepath.Base(repoRel) == name {
					filtered = append(filtered, repoRel)
					found = true
					break
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "  %s %s not found in repo\n", color.Yellow("warning"), name)
			}
		}
		entries = filtered
	}

	return linkEntries(r, entries, false)
}

// walkRepo returns all file paths in the repo relative to the repo root,
// skipping .git and other non-dotfile repo metadata.
func walkRepo(r *repo.Repo) ([]string, error) {
	var entries []string
	err := filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(r.Path, path)
		if err != nil {
			return err
		}
		// Skip repo root
		if rel == "." {
			return nil
		}
		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		// Skip non-dotfile repo metadata at the root level
		if !strings.Contains(rel, string(filepath.Separator)) {
			switch d.Name() {
			case "LICENSE", "NOTICES", "README.md", "Makefile":
				return nil
			}
		}
		// Only include files (not directories — we create dirs as needed)
		if !d.IsDir() {
			entries = append(entries, rel)
		}
		return nil
	})
	return entries, err
}

func linkEntries(r *repo.Repo, entries []string, quiet bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}
	return linkEntriesToHome(r, home, entries, quiet)
}

func linkEntriesToHome(r *repo.Repo, home string, entries []string, quiet bool) error {
	var linked, skipped, failed int
	for _, repoRel := range entries {
		repoPath := filepath.Join(r.Path, repoRel)
		linkPath := filepath.Join(home, repoRel)

		// Check if symlink already exists and is correct
		if managed, _ := dotfile.IsManaged(linkPath, r.Path); managed {
			if !quiet {
				fmt.Printf("  %s %s\n", color.Green("ok      "), color.Dim(repoRel))
			}
			skipped++
			continue
		}

		// Check if something else exists at the target
		if info, err := os.Lstat(linkPath); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				// Stale symlink — remove it
				os.Remove(linkPath)
			} else {
				fmt.Printf("  %s %s %s\n", color.Red("conflict"), color.Bold(repoRel), color.Dim("(file exists, not a symlink)"))
				failed++
				continue
			}
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(linkPath), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", color.Red("error   "), repoRel, err)
			failed++
			continue
		}

		// Create relative symlink
		if err := dotfile.CreateRelativeSymlink(linkPath, repoPath); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: %v\n", color.Red("error   "), repoRel, err)
			failed++
			continue
		}

		fmt.Printf("  %s %s %s %s\n",
			color.Green("linked  "),
			color.Bold(repoRel),
			color.Dim("->"),
			color.Cyan("~/"+repoRel))
		linked++
	}

	fmt.Printf("\n%s, %s, %s\n",
		color.Greenf("%d linked", linked),
		color.Dimf("%d already ok", skipped),
		color.Redf("%d failed", failed))
	if failed > 0 {
		return fmt.Errorf("%d entries failed", failed)
	}
	return nil
}
