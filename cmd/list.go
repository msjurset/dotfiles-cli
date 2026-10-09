package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func List(showAll bool) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	inRepo, _ := isInside(r.Path, cwd)
	if inRepo {
		entries, err := os.ReadDir(cwd)
		if err != nil {
			return err
		}
		ignorePatterns, _ := dotfile.LoadIgnorePatterns(r.Path)
		var listed int
		ignoredCount := 0
		for _, e := range entries {
			name := e.Name()
			if name == "." || name == ".." || name[0] != '.' {
				continue
			}
			isIgnored := dotfile.IsIgnored(name, ignorePatterns)
			if isIgnored && !showAll {
				ignoredCount++
				continue
			}
			info, err := os.Lstat(filepath.Join(cwd, name))
			if err != nil {
				continue
			}
			kind := "file"
			if info.IsDir() {
				kind = "dir "
			}
			listed++
			if isIgnored {
				fmt.Printf("  %s %s  %s %s\n",
					color.Dim("ignored  "),
					color.Dim(kind),
					color.Dim(name),
					color.Dim("(rule match)"))
			} else {
				fmt.Printf("  %s %s  %s\n",
					color.Green("managed  "),
					color.Dim(kind),
					color.Bold(name))
			}
		}
		if listed == 0 {
			fmt.Println("No dotfiles found in current directory.")
		}
		if ignoredCount > 0 && !showAll {
			fmt.Printf("\n%s\n", color.Dim(fmt.Sprintf("(%d dotfile(s) hidden by ignore rules; use -a or --all to show)", ignoredCount)))
		}
		return nil
	}

	entries, err := dotfile.Scan(cwd, r.Path)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No dotfiles found in current directory.")
		return nil
	}

	ignoredCount := 0
	var printed int
	for _, e := range entries {
		if !e.Managed && e.Ignored && !showAll {
			ignoredCount++
			continue
		}
		kind := "file"
		if e.IsDir {
			kind = "dir "
		}
		printed++
		if e.Managed {
			fmt.Printf("  %s %s  %s %s %s\n",
				color.Green("managed  "),
				color.Dim(kind),
				color.Bold(e.Name),
				color.Dim("->"),
				color.Cyan(".dotfiles/"+e.RepoRel))
		} else {
			if e.Ignored {
				fmt.Printf("  %s %s  %s %s\n",
					color.Dim("ignored  "),
					color.Dim(kind),
					color.Dim(e.Name),
					color.Dim("(rule match)"))
			} else {
				fmt.Printf("  %s %s  %s\n",
					color.Yellow("unmanaged"),
					color.Dim(kind),
					e.Name)
			}
		}
	}
	if printed == 0 {
		fmt.Println("No dotfiles found in current directory.")
	}
	if ignoredCount > 0 && !showAll {
		fmt.Printf("\n%s\n", color.Dim(fmt.Sprintf("(%d unmanaged dotfile(s) hidden by ignore rules; use -a or --all to show)", ignoredCount)))
	}
	return nil
}
