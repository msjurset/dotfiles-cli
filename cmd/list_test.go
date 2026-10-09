package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCmdListInRepo(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("tracked"), 0644)
	os.WriteFile(filepath.Join(repoDir, ".DS_Store"), []byte("noise"), 0644)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(repoDir)

	if err := List(false); err != nil {
		t.Fatalf("List(false) in repo failed: %v", err)
	}

	if err := List(true); err != nil {
		t.Fatalf("List(true) in repo failed: %v", err)
	}
}

func TestCmdListHome(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	homeDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(homeDir)

	os.WriteFile(filepath.Join(homeDir, ".DS_Store"), []byte("noise"), 0644)
	os.WriteFile(filepath.Join(homeDir, ".unmanaged_file"), []byte("data"), 0644)

	if err := List(false); err != nil {
		t.Fatalf("List(false) failed: %v", err)
	}

	if err := List(true); err != nil {
		t.Fatalf("List(true) failed: %v", err)
	}
}
