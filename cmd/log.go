package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

func Log(args []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	count := 10
	var filterPath string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-n" && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 {
				count = n
			}
			i++
		} else if strings.HasPrefix(arg, "-n") {
			if n, err := strconv.Atoi(arg[2:]); err == nil && n > 0 {
				count = n
			}
		} else if !strings.HasPrefix(arg, "-") {
			// Resolve file path to repo relative path if applicable
			absPath := arg
			if strings.HasPrefix(absPath, "~/") {
				absPath = filepath.Join(home, absPath[2:])
			} else if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(cwd, absPath)
			}
			absPath = filepath.Clean(absPath)

			if inside, rel := isInside(r.Path, absPath); inside {
				filterPath = rel
			} else if insideHome, relHome := isInside(home, absPath); insideHome {
				filterPath = relHome
			} else {
				filterPath = arg
			}
		}
	}

	commits, err := r.Log(count, filterPath)
	if err != nil {
		return fmt.Errorf("reading git log: %w", err)
	}

	if len(commits) == 0 {
		if filterPath != "" {
			fmt.Printf("No commit history found for %s.\n", filterPath)
		} else {
			fmt.Println("No commit history found.")
		}
		return nil
	}

	if filterPath != "" {
		fmt.Printf("%s %s (last %d commits):\n", color.Bold("History for"), color.Cyan(filterPath), len(commits))
	}

	for _, c := range commits {
		fmt.Printf("  %s  %s  %s  %s\n",
			color.Yellow(c.Hash),
			color.Dim(fmt.Sprintf("%-16s", "("+c.RelDate+")")),
			color.Cyan(fmt.Sprintf("%-14s", c.Author)),
			c.Subject)
	}
	return nil
}
