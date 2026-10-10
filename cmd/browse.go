package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

// CollectBrowseFiles returns all dotfiles and configs in the repo available for browsing.
func CollectBrowseFiles(repoPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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
		if base == ".DS_Store" || base == "README.md" || base == "LICENSE" || base == "dotfiles.1" || base == ".active_profile" {
			return nil
		}
		rel, err := filepath.Rel(repoPath, path)
		if err == nil {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// BuildFzfPreviewCmd builds the shell command for fzf's preview window showing git diffs or file contents.
func BuildFzfPreviewCmd(repoPath string) string {
	return fmt.Sprintf("cd '%s' && (git diff --color=always HEAD -- {} 2>/dev/null | grep -q . && (echo '=== Uncommitted Git Diff ===' && git diff --color=always HEAD -- {}) || (echo '=== File Content ===' && (which bat >/dev/null 2>&1 && bat --color=always --line-range :200 {} 2>/dev/null || head -n 200 {} 2>/dev/null || cat {})))", repoPath)
}

func Browse(query string, printOnly bool) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("fzf"); err != nil {
		fmt.Printf("%s fzf is required for interactive browsing.\n", color.Yellow("notice:"))
		fmt.Println("Install it using your package manager:")
		fmt.Println("  macOS:  brew install fzf")
		fmt.Println("  Debian: sudo apt install fzf")
		fmt.Println("  Arch:   sudo pacman -S fzf")
		return fmt.Errorf("fzf not found in PATH")
	}

	files, err := CollectBrowseFiles(r.Path)
	if err != nil {
		return fmt.Errorf("collecting files: %w", err)
	}
	if len(files) == 0 {
		fmt.Println("No dotfiles found in repository.")
		return nil
	}

	previewCmd := BuildFzfPreviewCmd(r.Path)
	args := []string{
		"--preview", previewCmd,
		"--preview-window", "right:60%",
		"--prompt", "dotfiles> ",
		"--height", "80%",
		"--reverse",
	}

	if query != "" {
		args = append(args, "--query", query)
	}

	if printOnly {
		args = append(args, "--header", "Enter: print path | Esc: exit")
	} else {
		args = append(args, "--header", "Enter: edit file | Esc: exit")
	}

	var outBuf bytes.Buffer
	fzfCmd := exec.Command("fzf", args...)
	fzfCmd.Dir = r.Path
	fzfCmd.Stdin = bytes.NewReader([]byte(strings.Join(files, "\n")))
	fzfCmd.Stdout = &outBuf
	fzfCmd.Stderr = os.Stderr

	if err := fzfCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Code 130 is user cancel (Esc / Ctrl-C), 1 is no match found / abort
			if exitErr.ExitCode() == 130 || exitErr.ExitCode() == 1 {
				return nil
			}
		}
		return err
	}

	selected := strings.TrimSpace(outBuf.String())
	if selected == "" {
		return nil
	}

	targetAbs := filepath.Join(r.Path, selected)
	if printOnly {
		fmt.Println(targetAbs)
		return nil
	}

	editor, err := FindEditor()
	if err != nil {
		return err
	}

	fmt.Printf("Editing %s with %s...\n", color.Bold(selected), color.Cyan(editor))
	if err := RunEditor(editor, targetAbs); err != nil {
		return err
	}

	out, err := r.StatusPorcelain(selected)
	if err == nil && out != "" {
		fmt.Printf("\n%s %s has uncommitted changes. Run '%s' to commit and push.\n",
			color.Yellow("~"), color.Bold(selected), color.Bold("dotfiles update"))
	}

	return nil
}
