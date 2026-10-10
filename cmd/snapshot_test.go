package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/msjurset/dotfiles/repo"
)

func initTestGitRepo(t *testing.T) *repo.Repo {
	t.Helper()
	dir := t.TempDir()
	r := &repo.Repo{Path: dir}

	cmds := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s: %v", args, out, err)
		}
	}

	// Create initial file & commit
	f := filepath.Join(dir, ".zshrc")
	os.WriteFile(f, []byte("export V=1\n"), 0644)
	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "initial").Run()

	return r
}

func TestRunSnapshot(t *testing.T) {
	r := initTestGitRepo(t)

	// 1. Create named snapshot
	if err := RunSnapshot(r, "v1-clean", false, ""); err != nil {
		t.Fatalf("RunSnapshot failed: %v", err)
	}

	// 2. Duplicate snapshot should fail
	if err := RunSnapshot(r, "v1-clean", false, ""); err == nil {
		t.Fatalf("expected error creating duplicate snapshot, got nil")
	}

	// 3. List snapshots
	if err := RunSnapshot(r, "", true, ""); err != nil {
		t.Fatalf("listing snapshots failed: %v", err)
	}

	// 4. Create default timestamped snapshot
	if err := RunSnapshot(r, "", false, ""); err != nil {
		t.Fatalf("creating timestamped snapshot failed: %v", err)
	}

	// 5. Delete snapshot
	if err := RunSnapshot(r, "", false, "v1-clean"); err != nil {
		t.Fatalf("deleting snapshot failed: %v", err)
	}

	// 6. Delete nonexistent snapshot should fail
	if err := RunSnapshot(r, "", false, "v1-clean"); err == nil {
		t.Fatalf("expected error deleting nonexistent snapshot, got nil")
	}
}

func TestRunRollback(t *testing.T) {
	r := initTestGitRepo(t)
	mockHome := t.TempDir()

	// Initial link
	entries, _ := walkRepo(r)
	linkEntriesToHome(r, mockHome, entries, true)

	// Create snapshot "stable"
	if err := RunSnapshot(r, "stable", false, ""); err != nil {
		t.Fatalf("creating snapshot failed: %v", err)
	}

	// Modify .zshrc and add a new file .bashrc
	os.WriteFile(filepath.Join(r.Path, ".zshrc"), []byte("export V=2-broken\n"), 0644)
	os.WriteFile(filepath.Join(r.Path, ".bashrc"), []byte("# new config\n"), 0644)
	exec.Command("git", "-C", r.Path, "add", ".").Run()
	exec.Command("git", "-C", r.Path, "commit", "-m", "broke zshrc and added bashrc").Run()

	// Update links
	entries, _ = walkRepo(r)
	linkEntriesToHome(r, mockHome, entries, true)

	// Verify broken state
	content, _ := os.ReadFile(filepath.Join(mockHome, ".zshrc"))
	if !strings.Contains(string(content), "V=2-broken") {
		t.Fatalf("expected V=2-broken before rollback")
	}

	// Run Rollback to "stable"
	if err := RunRollback(r, mockHome, "stable"); err != nil {
		t.Fatalf("RunRollback failed: %v", err)
	}

	// Verify .zshrc is restored to V=1
	content, _ = os.ReadFile(filepath.Join(r.Path, ".zshrc"))
	if !strings.Contains(string(content), "export V=1") {
		t.Fatalf("expected V=1 in repo after rollback, got: %s", string(content))
	}
	homeContent, _ := os.ReadFile(filepath.Join(mockHome, ".zshrc"))
	if !strings.Contains(string(homeContent), "export V=1") {
		t.Fatalf("expected V=1 in home after rollback, got: %s", string(homeContent))
	}

	// Verify .bashrc was removed from repo
	if _, err := os.Stat(filepath.Join(r.Path, ".bashrc")); !os.IsNotExist(err) {
		t.Fatalf("expected .bashrc to be removed from repo after rollback")
	}

	// Verify safety snapshot was created
	snapshots, err := r.ListSnapshots()
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}
	foundSafety := false
	for _, s := range snapshots {
		if strings.HasPrefix(s.Name, "pre-rollback-") {
			foundSafety = true
			break
		}
	}
	if !foundSafety {
		t.Fatalf("expected safety snapshot starting with pre-rollback- to be created")
	}
}
