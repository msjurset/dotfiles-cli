package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/msjurset/dotfiles/repo"
)

func setupAddTest(t *testing.T) (r *repo.Repo, home, cwd string) {
	t.Helper()
	repoDir := t.TempDir()
	home = t.TempDir()
	cwd = home
	r = &repo.Repo{Path: repoDir}
	return
}

func TestAddOneBasic(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	os.WriteFile(filepath.Join(home, ".zshrc"), []byte("# zsh"), 0644)

	if err := addOne(r, home, cwd, ".zshrc"); err != nil {
		t.Fatal(err)
	}

	// File should be in repo
	data, err := os.ReadFile(filepath.Join(r.Path, ".zshrc"))
	if err != nil {
		t.Fatal("file not moved to repo")
	}
	if string(data) != "# zsh" {
		t.Errorf("content = %q, want %q", string(data), "# zsh")
	}

	// Symlink should exist at original location
	info, err := os.Lstat(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatal("symlink not created")
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink")
	}
}

func TestAddOneNestedPath(t *testing.T) {
	r, home, _ := setupAddTest(t)

	// Create a nested dotfile: ~/work/project/.claude
	nested := filepath.Join(home, "work", "project", ".claude")
	os.MkdirAll(filepath.Dir(nested), 0755)
	os.WriteFile(nested, []byte("# claude"), 0644)

	// Add using an absolute path
	if err := addOne(r, home, home, nested); err != nil {
		t.Fatal(err)
	}

	// Should be stored at repo/work/project/.claude
	repoPath := filepath.Join(r.Path, "work", "project", ".claude")
	data, err := os.ReadFile(repoPath)
	if err != nil {
		t.Fatalf("file not at expected repo path %s: %v", repoPath, err)
	}
	if string(data) != "# claude" {
		t.Errorf("content = %q, want %q", string(data), "# claude")
	}

	// Symlink at original location
	info, err := os.Lstat(nested)
	if err != nil {
		t.Fatal("symlink not created")
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink")
	}
}

func TestAddOneCollision(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	// Pre-populate repo with a .claude
	os.WriteFile(filepath.Join(r.Path, ".claude"), []byte("# existing"), 0644)

	// Try to add a .claude from home
	os.WriteFile(filepath.Join(home, ".claude"), []byte("# new"), 0644)

	err := addOne(r, home, cwd, ".claude")
	if err == nil {
		t.Fatal("expected collision error")
	}
	if !contains(err.Error(), "collision") {
		t.Errorf("error = %q, want collision message", err.Error())
	}

	// Original file should still be in place (not moved)
	data, err := os.ReadFile(filepath.Join(home, ".claude"))
	if err != nil {
		t.Fatal("original file was removed despite collision")
	}
	if string(data) != "# new" {
		t.Errorf("original file was modified")
	}
}

func TestAddOneNotUnderHome(t *testing.T) {
	r, home, _ := setupAddTest(t)

	outside := t.TempDir()
	f := filepath.Join(outside, ".thing")
	os.WriteFile(f, []byte("x"), 0644)

	err := addOne(r, home, outside, ".thing")
	if err == nil {
		t.Fatal("expected error for path outside home")
	}
}

func TestAddOneNonDotfile(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	os.WriteFile(filepath.Join(home, "somefile"), []byte("data"), 0644)

	if err := addOne(r, home, cwd, "somefile"); err != nil {
		t.Fatalf("non-dotfile add should work: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(r.Path, "somefile")); err != nil {
		t.Fatal("file not in repo")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
