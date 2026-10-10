package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectBrowseFiles(t *testing.T) {
	repoDir := t.TempDir()

	// Create structure
	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("export FOO=1\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, ".aliases"), []byte("alias ll='ls -la'\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Dotfiles\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, "LICENSE"), []byte("MIT\n"), 0644)
	os.WriteFile(filepath.Join(repoDir, ".DS_Store"), []byte("dummy"), 0644)
	os.WriteFile(filepath.Join(repoDir, ".active_profile"), []byte("work"), 0644)

	configDir := filepath.Join(repoDir, ".config", "kitty")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "kitty.conf"), []byte("font_size 12.0\n"), 0644)

	gitDir := filepath.Join(repoDir, ".git", "objects")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "ignore_me"), []byte("git object"), 0644)

	files, err := CollectBrowseFiles(repoDir)
	if err != nil {
		t.Fatalf("CollectBrowseFiles failed: %v", err)
	}

	expected := map[string]bool{
		".aliases":                 true,
		".zshrc":                   true,
		filepath.Join(".config", "kitty", "kitty.conf"): true,
	}

	if len(files) != len(expected) {
		t.Fatalf("expected %d files, got %d: %v", len(expected), len(files), files)
	}

	for _, f := range files {
		if !expected[f] {
			t.Errorf("unexpected file in browse list: %s", f)
		}
	}
}

func TestBuildFzfPreviewCmd(t *testing.T) {
	cmdStr := BuildFzfPreviewCmd("/path/to/repo")
	if cmdStr == "" {
		t.Fatalf("expected non-empty preview command")
	}
}
