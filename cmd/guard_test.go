package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdGuard(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	// List guards
	if err := Guard(nil); err != nil {
		t.Fatalf("Guard(nil) failed: %v", err)
	}

	// Add pattern
	if err := Guard([]string{"id_custom*", "*.keypair"}); err != nil {
		t.Fatalf("Guard add failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(repoDir, ".dotfilesguard"))
	if err != nil {
		t.Fatalf("reading .dotfilesguard: %v", err)
	}
	if !strings.Contains(string(content), "id_custom*") || !strings.Contains(string(content), "*.keypair") {
		t.Errorf("expected patterns in .dotfilesguard, got: %s", string(content))
	}

	// Add duplicate pattern
	if err := Guard([]string{"id_custom*"}); err != nil {
		t.Fatalf("Guard duplicate failed: %v", err)
	}
}
