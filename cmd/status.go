package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Status() error {
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

	var managed, unmanaged []dotfile.Entry
	var repoPaths []string
	for _, e := range entries {
		if e.Managed {
			managed = append(managed, e)
			repoPaths = append(repoPaths, e.RepoRel)
		} else {
			unmanaged = append(unmanaged, e)
		}
	}

	// Check git status for managed files
	if len(repoPaths) > 0 {
		out, err := r.StatusPorcelain(repoPaths...)
		if err != nil {
			return err
		}
		if out != "" {
			fmt.Println(color.Bold("Changes:"))
			for _, line := range strings.Split(out, "\n") {
				if line != "" {
					fmt.Printf("  %s\n", colorStatus(line))
				}
			}
			fmt.Println()
		}
	}

	// Check unpushed commits
	n, err := r.UnpushedCount()
	if err != nil {
		return err
	}
	if n > 0 {
		fmt.Printf("%s\n\n", color.Yellowf("%d unpushed commit(s)", n))
	}

	if len(managed) == 0 && len(unmanaged) == 0 {
		fmt.Println("No dotfiles found in current directory.")
		return nil
	}

	if len(managed) > 0 {
		out, _ := r.StatusPorcelain(repoPaths...)
		if out == "" && n == 0 {
			fmt.Println(color.Green("All managed dotfiles are clean."))
		}
	}

	if len(unmanaged) > 0 {
		fmt.Println(color.Bold("Unmanaged dotfiles:"))
		for _, e := range unmanaged {
			kind := "file"
			if e.IsDir {
				kind = "dir "
			}
			fmt.Printf("  %s  %s\n", color.Dim(kind), color.Yellow(e.Name))
		}
	}
	return nil
}

func colorStatus(line string) string {
	if len(line) < 3 {
		return line
	}
	code := line[:2]
	rest := line[2:]
	switch {
	case strings.ContainsAny(code, "MADRC"):
		return color.Green(code) + color.Cyan(rest)
	case strings.Contains(code, "?"):
		return color.Yellow(code) + rest
	case strings.Contains(code, "D"):
		return color.Red(code) + rest
	default:
		return line
	}
}
