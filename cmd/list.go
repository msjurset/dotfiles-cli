package cmd

import (
	"fmt"
	"os"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func List() error {
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
	if len(entries) == 0 {
		fmt.Println("No dotfiles found in current directory.")
		return nil
	}
	for _, e := range entries {
		kind := "file"
		if e.IsDir {
			kind = "dir "
		}
		if e.Managed {
			fmt.Printf("  %s %s  %s %s %s\n",
				color.Green("managed  "),
				color.Dim(kind),
				color.Bold(e.Name),
				color.Dim("->"),
				color.Cyan(".dotfiles/"+e.RepoRel))
		} else {
			fmt.Printf("  %s %s  %s\n",
				color.Yellow("unmanaged"),
				color.Dim(kind),
				e.Name)
		}
	}
	return nil
}
