package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCmdLog(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	if out, err := exec.Command("git", "-C", repoDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %s: %v", out, err)
	}
	_ = exec.Command("git", "-C", repoDir, "config", "user.name", "Test").Run()
	_ = exec.Command("git", "-C", repoDir, "config", "user.email", "test@example.com").Run()

	testFile := filepath.Join(repoDir, ".zshrc")
	os.WriteFile(testFile, []byte("alias ll='ls -la'\n"), 0644)
	_ = exec.Command("git", "-C", repoDir, "add", ".zshrc").Run()
	_ = exec.Command("git", "-C", repoDir, "commit", "-m", "init zshrc").Run()

	if err := Log(nil); err != nil {
		t.Fatalf("Log(nil) failed: %v", err)
	}

	if err := Log([]string{"-n", "1"}); err != nil {
		t.Fatalf("Log(-n 1) failed: %v", err)
	}

	if err := Log([]string{".zshrc"}); err != nil {
		t.Fatalf("Log(.zshrc) failed: %v", err)
	}
}
