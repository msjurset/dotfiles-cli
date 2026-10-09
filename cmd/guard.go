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

func Guard(patterns []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	if len(patterns) == 0 {
		return listGuards(r.Path)
	}

	added, err := dotfile.AppendGuardPattern(r.Path, patterns...)
	if err != nil {
		return err
	}

	if len(added) == 0 {
		fmt.Println("All patterns are already present in .dotfilesguard.")
		return nil
	}

	_ = r.Add(".dotfilesguard")

	for _, p := range added {
		fmt.Printf("  %s %s %s\n",
			color.Green("guarded"),
			color.Bold(p),
			color.Dim("(added to .dotfilesguard)"))
	}
	return nil
}

func listGuards(repoPath string) error {
	guardFile := filepath.Join(repoPath, ".dotfilesguard")
	var custom []string

	if f, err := os.Open(guardFile); err == nil {
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
		fmt.Println(color.Bold("Custom sensitive patterns (.dotfilesguard):"))
		for _, p := range custom {
			fmt.Printf("  %s\n", color.Red(p))
		}
		fmt.Println()
	}

	fmt.Println(color.Bold("Default sensitive guardrails:"))
	for _, p := range dotfile.DefaultGuardrails {
		fmt.Printf("  %s\n", color.Yellow(p))
	}

	return nil
}
