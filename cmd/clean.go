package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Clean(names []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	entries, err := dotfile.Scan(cwd, r.Path)
	if err != nil {
		return err
	}

	// Filter to managed entries only
	var targets []dotfile.Entry
	if len(names) > 0 {
		byName := make(map[string]dotfile.Entry)
		for _, e := range entries {
			if e.Managed {
				byName[e.Name] = e
			}
		}
		for _, name := range names {
			e, ok := byName[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "  %s %s is not managed, skipping\n", color.Yellow("warning"), name)
				continue
			}
			targets = append(targets, e)
		}
	} else {
		for _, e := range entries {
			if e.Managed {
				targets = append(targets, e)
			}
		}
	}

	if len(targets) == 0 {
		fmt.Println("No managed dotfiles to clean.")
		return nil
	}

	var cleaned int
	for _, e := range targets {
		repoPath := filepath.Join(r.Path, e.RepoRel)

		// Remove the symlink
		if err := os.Remove(e.AbsPath); err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s: removing symlink: %v\n", color.Red("error"), e.Name, err)
			continue
		}

		// Copy the file back from repo to original location
		if err := copyPath(repoPath, e.AbsPath); err != nil {
			// Attempt to restore symlink on failure
			_ = dotfile.CreateRelativeSymlink(e.AbsPath, repoPath)
			fmt.Fprintf(os.Stderr, "  %s %s: restoring file: %v\n", color.Red("error"), e.Name, err)
			continue
		}

		fmt.Printf("  %s %s %s\n",
			color.Green("cleaned"),
			color.Bold(e.Name),
			color.Dim("(restored from repo)"))
		cleaned++
	}

	fmt.Printf("\n%s\n", color.Greenf("%d dotfile(s) cleaned", cleaned))
	return nil
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPath(filepath.Join(src, e.Name()), filepath.Join(dst, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
