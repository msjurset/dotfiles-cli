package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCmdDiff(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	// Init git repo
	if out, err := exec.Command("git", "-C", repoDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	_ = exec.Command("git", "-C", repoDir, "config", "user.name", "Test").Run()
	_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()

	testFile := filepath.Join(repoDir, ".testfile")
	os.WriteFile(testFile, []byte("line1\n"), 0644)
	_ = exec.Command("git", "-C", repoDir, "add", ".testfile").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "init").Run()

	// Modify file
	os.WriteFile(testFile, []byte("line1\nline2\n"), 0644)

	// Diff working tree
	if err := Diff(nil); err != nil {
		t.Fatalf("Diff unstaged failed: %v", err)
	}

	// Stage change
	_ = exec.Command("git", "-C", repoDir, "add", ".testfile").Run()

	// Diff staged
	if err := Diff([]string{"--staged"}); err != nil {
		t.Fatalf("Diff staged failed: %v", err)
	}
}
