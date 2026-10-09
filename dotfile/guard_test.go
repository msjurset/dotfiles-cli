package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGuard(t *testing.T) {
	tempDir := t.TempDir()

	patterns, err := LoadGuardPatterns(tempDir)
	if err != nil {
		t.Fatalf("LoadGuardPatterns failed: %v", err)
	}

	tests := []struct {
		path      string
		sensitive bool
	}{
		{".ssh", true},
		{".ssh/id_rsa", true},
		{".ssh/id_ed25519", true},
		{"id_rsa", true},
		{"cert.pem", true},
		{"server.key", true},
		{".secrets", true},
		{".env.production", true},
		{".aws", true},
		{".zshrc", false},
		{".aliases", false},
		{".vimrc", false},
	}

	for _, tt := range tests {
		got, _ := IsSensitive(tt.path, patterns)
		if got != tt.sensitive {
			t.Errorf("IsSensitive(%q) = %v, want %v", tt.path, got, tt.sensitive)
		}
	}

	// Test appending custom guard pattern
	added, err := AppendGuardPattern(tempDir, "*.kdbx", ".private_tokens*")
	if err != nil {
		t.Fatalf("AppendGuardPattern failed: %v", err)
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 added, got %d", len(added))
	}

	patterns, err = LoadGuardPatterns(tempDir)
	if err != nil {
		t.Fatalf("LoadGuardPatterns failed: %v", err)
	}
	if sensitive, _ := IsSensitive("passwords.kdbx", patterns); !sensitive {
		t.Errorf("expected passwords.kdbx to be sensitive")
	}

	_ = os.Remove(filepath.Join(tempDir, ".dotfilesguard"))
}
