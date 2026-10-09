package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/msjurset/dotfiles/dotfile"
)

func TestDoctorClean(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)
	t.Setenv("HOME", homeDir)

	// Create repo file and proper symlink in home
	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("# zsh"), 0644)
	if err := dotfile.CreateRelativeSymlink(filepath.Join(homeDir, ".zshrc"), filepath.Join(repoDir, ".zshrc")); err != nil {
		t.Fatal(err)
	}

	if err := Doctor(false); err != nil {
		t.Fatalf("expected doctor clean, got error: %v", err)
	}
}

func TestDoctorClobberedAndMissing(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)
	t.Setenv("HOME", homeDir)

	// 1. Missing symlink
	os.WriteFile(filepath.Join(repoDir, ".missing_link"), []byte("# missing"), 0644)

	// 2. Clobbered symlink (regular file in home instead of symlink)
	os.WriteFile(filepath.Join(repoDir, ".clobbered"), []byte("repo version"), 0644)
	os.WriteFile(filepath.Join(homeDir, ".clobbered"), []byte("local modified version"), 0644)

	// Doctor check without fix
	if err := Doctor(false); err != nil {
		t.Fatalf("Doctor(false) failed: %v", err)
	}

	// Now fix
	if err := Doctor(true); err != nil {
		t.Fatalf("Doctor(true) failed: %v", err)
	}

	// Verify missing link was created
	info, err := os.Lstat(filepath.Join(homeDir, ".missing_link"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected .missing_link to be a symlink")
	}

	// Verify clobbered link was restored as symlink
	info, err = os.Lstat(filepath.Join(homeDir, ".clobbered"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected .clobbered to be a symlink after fix")
	}

	// Verify doctor passes cleanly now
	if err := Doctor(false); err != nil {
		t.Fatalf("Doctor after fix should be clean, got: %v", err)
	}
}
