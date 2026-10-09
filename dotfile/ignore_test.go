package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnore(t *testing.T) {
	tempDir := t.TempDir()

	patterns, err := LoadIgnorePatterns(tempDir)
	if err != nil {
		t.Fatalf("LoadIgnorePatterns failed: %v", err)
	}

	// Test default ignores
	tests := []struct {
		name     string
		expected bool
	}{
		{".git", true},
		{".DS_Store", true},
		{".Trash", true},
		{".zsh_history", true},
		{".bash_history", true},
		{".cache", true},
		{".npm", true},
		{".cargo", true},
		{".zshrc", false},
		{".aliases", false},
		{".gitconfig", false},
	}

	for _, tt := range tests {
		got := IsIgnored(tt.name, patterns)
		if got != tt.expected {
			t.Errorf("IsIgnored(%q) = %v, want %v", tt.name, got, tt.expected)
		}
	}

	// Test appending custom patterns
	added, err := AppendIgnorePattern(tempDir, ".custom_ignore", ".another_test")
	if err != nil {
		t.Fatalf("AppendIgnorePattern failed: %v", err)
	}
	if len(added) != 2 {
		t.Fatalf("expected 2 added, got %d", len(added))
	}

	patterns, err = LoadIgnorePatterns(tempDir)
	if err != nil {
		t.Fatalf("LoadIgnorePatterns failed: %v", err)
	}
	if !IsIgnored(".custom_ignore", patterns) {
		t.Errorf("expected .custom_ignore to be ignored")
	}

	// Test duplicate append does not re-add
	added2, err := AppendIgnorePattern(tempDir, ".custom_ignore")
	if err != nil {
		t.Fatalf("AppendIgnorePattern duplicate failed: %v", err)
	}
	if len(added2) != 0 {
		t.Errorf("expected 0 added for duplicate, got %d", len(added2))
	}

	_ = os.Remove(filepath.Join(tempDir, ".dotfilesignore"))
}
