package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdIgnore(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	// List ignores (should not error when empty)
	if err := Ignore(nil); err != nil {
		t.Fatalf("Ignore(nil) failed: %v", err)
	}

	// Add pattern
	if err := Ignore([]string{".mycustom_ignore"}); err != nil {
		t.Fatalf("Ignore add failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(repoDir, ".dotfilesignore"))
	if err != nil {
		t.Fatalf("reading .dotfilesignore: %v", err)
	}
	if !strings.Contains(string(content), ".mycustom_ignore") {
		t.Errorf("expected pattern in .dotfilesignore, got: %s", string(content))
	}

	// Add duplicate pattern
	if err := Ignore([]string{".mycustom_ignore"}); err != nil {
		t.Fatalf("Ignore duplicate failed: %v", err)
	}
}
