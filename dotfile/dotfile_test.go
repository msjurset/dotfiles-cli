package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()

	// Create a managed dotfile: file in repo, symlink in home
	repoFile := filepath.Join(repoDir, ".zshrc")
	os.WriteFile(repoFile, []byte("# zsh"), 0644)
	rel, _ := filepath.Rel(homeDir, repoFile)
	os.Symlink(rel, filepath.Join(homeDir, ".zshrc"))

	// Create an unmanaged dotfile
	os.WriteFile(filepath.Join(homeDir, ".bashrc"), []byte("# bash"), 0644)

	// Create an unmanaged dotdir
	os.Mkdir(filepath.Join(homeDir, ".config"), 0755)

	// Create a non-dot file (should be ignored)
	os.WriteFile(filepath.Join(homeDir, "readme.txt"), []byte("hi"), 0644)

	entries, err := Scan(homeDir, repoDir)
	if err != nil {
		t.Fatal(err)
	}

	// Build lookup
	byName := make(map[string]Entry)
	for _, e := range entries {
		byName[e.Name] = e
	}

	tests := []struct {
		name    string
		wantDir bool
		wantMgd bool
	}{
		{".zshrc", false, true},
		{".bashrc", false, false},
		{".config", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, ok := byName[tt.name]
			if !ok {
				t.Fatalf("%s not found in scan results", tt.name)
			}
			if e.IsDir != tt.wantDir {
				t.Errorf("IsDir = %v, want %v", e.IsDir, tt.wantDir)
			}
			if e.Managed != tt.wantMgd {
				t.Errorf("Managed = %v, want %v", e.Managed, tt.wantMgd)
			}
		})
	}

	if _, ok := byName["readme.txt"]; ok {
		t.Error("non-dotfile should not appear in results")
	}
}
