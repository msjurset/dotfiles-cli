package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuditClean(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("alias ll='ls -la'\n"), 0644)

	if err := Audit(nil); err != nil {
		t.Fatalf("expected clean audit, got: %v", err)
	}
}

func TestAuditWithSecret(t *testing.T) {
	repoDir := t.TempDir()
	t.Setenv("DOTFILES_REPO_PATH", repoDir)

	secretFile := filepath.Join(repoDir, ".secret_config")
	os.WriteFile(secretFile, []byte("export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\n"), 0644)

	err := Audit(nil)
	if err == nil {
		t.Fatal("expected audit to detect secret, got nil")
	}

	// Test CheckPathsForSecrets without force
	err = CheckPathsForSecrets(repoDir, []string{".secret_config"}, false)
	if err == nil {
		t.Fatal("expected CheckPathsForSecrets to block without force")
	}

	// Test CheckPathsForSecrets with force
	err = CheckPathsForSecrets(repoDir, []string{".secret_config"}, true)
	if err != nil {
		t.Fatalf("expected CheckPathsForSecrets to pass with force, got: %v", err)
	}
}
