package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Ignore(patterns []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	if len(patterns) == 0 {
		return listIgnores(r.Path)
	}

	added, err := dotfile.AppendIgnorePattern(r.Path, patterns...)
	if err != nil {
		return err
	}

	if len(added) == 0 {
		fmt.Println("All patterns are already present in .dotfilesignore.")
		return nil
	}

	_ = r.Add(".dotfilesignore")

	for _, p := range added {
		fmt.Printf("  %s %s %s\n",
			color.Green("ignored"),
			color.Bold(p),
			color.Dim("(added to .dotfilesignore)"))
	}
	return nil
}

func listIgnores(repoPath string) error {
	ignoreFile := filepath.Join(repoPath, ".dotfilesignore")
	var custom []string

	if f, err := os.Open(ignoreFile); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				custom = append(custom, line)
			}
		}
	}

	if len(custom) > 0 {
		fmt.Println(color.Bold("Custom ignore patterns (.dotfilesignore):"))
		for _, p := range custom {
			fmt.Printf("  %s\n", color.Cyan(p))
		}
		fmt.Println()
	}

	fmt.Println(color.Bold("Default ignore patterns:"))
	for _, p := range dotfile.DefaultIgnores {
		fmt.Printf("  %s\n", color.Dim(p))
	}

	return nil
}
