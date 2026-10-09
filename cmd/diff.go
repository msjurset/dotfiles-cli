package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

func Diff(args []string) error {
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

	staged := false
	var paths []string

	for _, arg := range args {
		switch arg {
		case "--staged", "--cached":
			staged = true
		default:
			// Resolve relative or home path to repo-relative path if applicable
			absPath := arg
			if strings.HasPrefix(absPath, "~/") {
				absPath = filepath.Join(home, absPath[2:])
			} else if !filepath.IsAbs(absPath) {
				absPath = filepath.Join(cwd, absPath)
			}
			absPath = filepath.Clean(absPath)

			// If inside repo, get repo-relative path
			if inside, rel := isInside(r.Path, absPath); inside {
				paths = append(paths, rel)
			} else if insideHome, relHome := isInside(home, absPath); insideHome {
				// Path was in home, map to repo path
				paths = append(paths, relHome)
			} else {
				paths = append(paths, arg)
			}
		}
	}

	out, err := r.Diff(staged, paths...)
	if err != nil {
		return err
	}

	if out == "" {
		if staged {
			fmt.Println(color.Green("No staged changes."))
		} else {
			fmt.Println(color.Green("No working tree changes."))
		}
		return nil
	}

	fmt.Println(out)
	return nil
}
