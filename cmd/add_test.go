package cmd

import (
	"os"
	"os/exec"
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

	if err := addOne(r, home, cwd, ".zshrc", false); err != nil {
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
	if err := addOne(r, home, home, nested, false); err != nil {
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

	err := addOne(r, home, cwd, ".claude", false)
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

	err := addOne(r, home, outside, ".thing", false)
	if err == nil {
		t.Fatal("expected error for path outside home")
	}
}

func TestAddOneNonDotfile(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	os.WriteFile(filepath.Join(home, "somefile"), []byte("data"), 0644)

	if err := addOne(r, home, cwd, "somefile", false); err != nil {
		t.Fatalf("non-dotfile add should work: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(r.Path, "somefile")); err != nil {
		t.Fatal("file not in repo")
	}
}

func TestAddOneInsideManagedSymlinkedDir(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	// Initialize git repo so r.Add() works
	if out, err := exec.Command("git", "-C", r.Path, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}

	// Create a directory in the repo with an existing file
	zfuncsRepo := filepath.Join(r.Path, ".zfuncs")
	os.MkdirAll(zfuncsRepo, 0755)
	os.WriteFile(filepath.Join(zfuncsRepo, "existing"), []byte("# existing"), 0644)

	// Symlink home/.zfuncs -> repo/.zfuncs (simulating a managed directory)
	os.Symlink(zfuncsRepo, filepath.Join(home, ".zfuncs"))

	// Create a new file directly in the repo dir (user added it by hand)
	os.WriteFile(filepath.Join(zfuncsRepo, "newfile"), []byte("# new func"), 0644)

	// Add via the symlinked path — should detect it's already in the repo
	err := addOne(r, home, cwd, ".zfuncs/newfile", false)
	if err != nil {
		t.Fatalf("expected success for file already in repo, got: %v", err)
	}

	// File should still be a regular file at the repo path (not moved elsewhere)
	info, err := os.Lstat(filepath.Join(zfuncsRepo, "newfile"))
	if err != nil {
		t.Fatal("file should still exist in repo")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("file in repo should not be a symlink")
	}

	// Content unchanged
	data, err := os.ReadFile(filepath.Join(zfuncsRepo, "newfile"))
	if err != nil {
		t.Fatal("file should be readable")
	}
	if string(data) != "# new func" {
		t.Errorf("content = %q, want %q", string(data), "# new func")
	}
}

func TestAddOneSensitiveGuardrails(t *testing.T) {
	r, home, cwd := setupAddTest(t)

	// Create sensitive file
	envFile := filepath.Join(home, ".env.production")
	os.WriteFile(envFile, []byte("SECRET=123"), 0600)

	// Attempt without force should fail
	err := addOne(r, home, cwd, ".env.production", false)
	if err == nil {
		t.Fatal("expected sensitive guardrail error, got nil")
	}
	if !contains(err.Error(), "sensitive guardrail") {
		t.Fatalf("expected sensitive guardrail error, got: %v", err)
	}

	// Original file must still exist and not be moved
	if _, err := os.Stat(envFile); err != nil {
		t.Fatal("original file was removed")
	}

	// Attempt with force should succeed
	if err := addOne(r, home, cwd, ".env.production", true); err != nil {
		t.Fatalf("forced add failed: %v", err)
	}

	// Check it was moved and symlinked
	data, err := os.ReadFile(filepath.Join(r.Path, ".env.production"))
	if err != nil || string(data) != "SECRET=123" {
		t.Fatal("file not moved to repo")
	}

	info, err := os.Lstat(envFile)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink not created")
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
