package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", dir)

	r, err := Discover()
	if err != nil {
		t.Fatal(err)
	}
	if r.Path != dir {
		t.Errorf("Path = %q, want %q", r.Path, dir)
	}
}

func TestDiscoverMissingDir(t *testing.T) {
	t.Setenv("DOTFILES_REPO_PATH", "/tmp/nonexistent-dotfiles-repo-test")

	_, err := Discover()
	if err == nil {
		t.Fatal("expected error for missing repo dir")
	}
}

func TestDiscoverNotADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "not-a-dir")
	os.WriteFile(f, []byte("hello"), 0644)
	t.Setenv("DOTFILES_REPO_PATH", f)

	_, err := Discover()
	if err == nil {
		t.Fatal("expected error when path is a file, not directory")
	}
}

func TestDiscoverFallback(t *testing.T) {
	// Unset the env var so it falls back to ~/.dotfiles
	t.Setenv("DOTFILES_REPO_PATH", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}
	dotfilesDir := filepath.Join(home, ".dotfiles")
	info, err := os.Stat(dotfilesDir)
	if err != nil || !info.IsDir() {
		t.Skip("~/.dotfiles does not exist, skipping fallback test")
	}

	r, err := Discover()
	if err != nil {
		t.Fatal(err)
	}
	if r.Path != dotfilesDir {
		t.Errorf("Path = %q, want %q", r.Path, dotfilesDir)
	}
}
