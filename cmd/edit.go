package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

func Edit(query string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	editor, err := FindEditor()
	if err != nil {
		return err
	}

	// If no query passed, open repo directory
	if query == "" {
		return RunEditor(editor, r.Path)
	}

	// Search for dotfile matching query
	var allFiles []string
	err = filepath.Walk(r.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		base := info.Name()
		if base == "README.md" || base == "LICENSE" {
			return nil
		}
		rel, err := filepath.Rel(r.Path, path)
		if err == nil {
			allFiles = append(allFiles, rel)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking repo: %w", err)
	}

	target := matchDotfile(query, allFiles)
	if target == "" {
		return fmt.Errorf("no managed dotfile matching %q (use 'dotfiles list' to see tracked files)", query)
	}

	targetAbs := filepath.Join(r.Path, target)
	fmt.Printf("Editing %s with %s...\n", color.Bold(target), color.Cyan(editor))
	if err := RunEditor(editor, targetAbs); err != nil {
		return err
	}

	// Check if target was modified
	out, err := r.StatusPorcelain(target)
	if err == nil && out != "" {
		fmt.Printf("\n%s %s has uncommitted changes. Run '%s' to commit and push.\n",
			color.Yellow("~"), color.Bold(target), color.Bold("dotfiles update"))
	}

	return nil
}

func matchDotfile(query string, files []string) string {
	q := strings.ToLower(strings.TrimSpace(query))
	// 1. Exact match (e.g. ".zshrc" or "zshrc")
	for _, f := range files {
		base := strings.ToLower(filepath.Base(f))
		rel := strings.ToLower(f)
		trimmedBase := strings.TrimPrefix(base, ".")
		if base == q || rel == q || trimmedBase == q {
			return f
		}
	}

	// 2. Prefix match (e.g. "zsh" matches ".zshrc")
	var prefixMatches []string
	for _, f := range files {
		base := strings.ToLower(filepath.Base(f))
		trimmedBase := strings.TrimPrefix(base, ".")
		if strings.HasPrefix(base, q) || strings.HasPrefix(trimmedBase, q) {
			prefixMatches = append(prefixMatches, f)
		}
	}
	if len(prefixMatches) == 1 {
		return prefixMatches[0]
	}

	// 3. Substring match
	var subMatches []string
	for _, f := range files {
		if strings.Contains(strings.ToLower(f), q) {
			subMatches = append(subMatches, f)
		}
	}
	if len(subMatches) == 1 {
		return subMatches[0]
	}

	if len(prefixMatches) > 0 {
		return prefixMatches[0]
	}
	if len(subMatches) > 0 {
		return subMatches[0]
	}

	return ""
}

func FindEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, candidate := range []string{"nvim", "vim", "vi", "nano"} {
			if _, err := exec.LookPath(candidate); err == nil {
				return candidate, nil
			}
		}
	}
	if editor == "" {
		return "", fmt.Errorf("no editor found (set $EDITOR or $VISUAL)")
	}
	return editor, nil
}

func RunEditor(editor string, path string) error {
	parts := strings.Fields(editor)
	args := append(parts[1:], path)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
