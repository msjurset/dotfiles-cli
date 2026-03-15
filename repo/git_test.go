package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) *Repo {
	t.Helper()
	dir := t.TempDir()
	r := &Repo{Path: dir}

	// Initialize a git repo with an initial commit
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

	// Create initial commit
	f := filepath.Join(dir, ".gitkeep")
	os.WriteFile(f, []byte(""), 0644)
	exec.Command("git", "-C", dir, "add", ".").Run()
	exec.Command("git", "-C", dir, "commit", "-m", "init").Run()

	return r
}

func TestGit(t *testing.T) {
	r := initTestRepo(t)

	out, err := r.Git("status", "--porcelain")
	if err != nil {
		t.Fatalf("git status failed: %v", err)
	}
	if out != "" {
		t.Errorf("expected clean status, got %q", out)
	}
}

func TestGitInvalidCommand(t *testing.T) {
	r := initTestRepo(t)
	_, err := r.Git("not-a-real-command")
	if err == nil {
		t.Fatal("expected error for invalid git command")
	}
}

func TestStatusPorcelain(t *testing.T) {
	r := initTestRepo(t)

	// Create an untracked file
	os.WriteFile(filepath.Join(r.Path, ".zshrc"), []byte("# zsh"), 0644)

	out, err := r.StatusPorcelain()
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected non-empty status with untracked file")
	}
}

func TestStatusPorcelainScoped(t *testing.T) {
	r := initTestRepo(t)

	os.WriteFile(filepath.Join(r.Path, ".zshrc"), []byte("# zsh"), 0644)
	os.WriteFile(filepath.Join(r.Path, ".bashrc"), []byte("# bash"), 0644)

	out, err := r.StatusPorcelain(".zshrc")
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected status for .zshrc")
	}
	// Should only contain .zshrc, not .bashrc
	if len(out) > 0 {
		for _, line := range splitLines(out) {
			if line != "" && !contains(line, ".zshrc") {
				t.Errorf("unexpected file in scoped status: %q", line)
			}
		}
	}
}

func TestAddAndCommit(t *testing.T) {
	r := initTestRepo(t)

	os.WriteFile(filepath.Join(r.Path, ".vimrc"), []byte("set nocp"), 0644)

	if err := r.Add(".vimrc"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if err := r.Commit("add vimrc"); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify clean after commit
	out, err := r.StatusPorcelain()
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Errorf("expected clean status after commit, got %q", out)
	}
}

func TestUnpushedCountNoUpstream(t *testing.T) {
	r := initTestRepo(t)

	n, err := r.UnpushedCount()
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("expected 0 unpushed (no upstream), got %d", n)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
