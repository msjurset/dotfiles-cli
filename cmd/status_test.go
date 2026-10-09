package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCmdStatusInRepo(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	if out, err := exec.Command("git", "-C", repoDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(repoDir)

	// Status when in repo
	if err := Status(false); err != nil {
		t.Fatalf("Status(false) in repo failed: %v", err)
	}
}

func TestCmdStatusHome(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	if out, err := exec.Command("git", "-C", repoDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}

	homeDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	_ = os.Chdir(homeDir)

	// Create ignored file and normal unmanaged file
	os.WriteFile(filepath.Join(homeDir, ".DS_Store"), []byte("noise"), 0644)
	os.WriteFile(filepath.Join(homeDir, ".myconfig"), []byte("active"), 0644)

	if err := Status(false); err != nil {
		t.Fatalf("Status(false) failed: %v", err)
	}

	if err := Status(true); err != nil {
		t.Fatalf("Status(true) failed: %v", err)
	}
}
