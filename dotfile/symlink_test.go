package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsManaged(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()

	// Create a file in the repo
	repoFile := filepath.Join(repoDir, ".zshrc")
	if err := os.WriteFile(repoFile, []byte("# zshrc"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantManag bool
		wantRel   string
	}{
		{
			name: "symlink into repo",
			setup: func(t *testing.T) string {
				link := filepath.Join(homeDir, ".zshrc")
				rel, _ := filepath.Rel(homeDir, repoFile)
				os.Symlink(rel, link)
				t.Cleanup(func() { os.Remove(link) })
				return link
			},
			wantManag: true,
			wantRel:   ".zshrc",
		},
		{
			name: "symlink outside repo",
			setup: func(t *testing.T) string {
				otherDir := t.TempDir()
				otherFile := filepath.Join(otherDir, ".bashrc")
				os.WriteFile(otherFile, []byte("# bash"), 0644)
				link := filepath.Join(homeDir, ".bashrc")
				os.Symlink(otherFile, link)
				t.Cleanup(func() { os.Remove(link) })
				return link
			},
			wantManag: false,
		},
		{
			name: "regular file",
			setup: func(t *testing.T) string {
				f := filepath.Join(homeDir, ".localrc")
				os.WriteFile(f, []byte("# local"), 0644)
				t.Cleanup(func() { os.Remove(f) })
				return f
			},
			wantManag: false,
		},
		{
			name: "nonexistent",
			setup: func(t *testing.T) string {
				return filepath.Join(homeDir, ".nope")
			},
			wantManag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			managed, rel := IsManaged(path, repoDir)
			if managed != tt.wantManag {
				t.Errorf("managed = %v, want %v", managed, tt.wantManag)
			}
			if tt.wantManag && rel != tt.wantRel {
				t.Errorf("rel = %q, want %q", rel, tt.wantRel)
			}
		})
	}
}

func TestCreateRelativeSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "repo", ".vimrc")
	os.MkdirAll(filepath.Dir(target), 0755)
	os.WriteFile(target, []byte("\" vim"), 0644)

	link := filepath.Join(dir, "home", ".vimrc")
	os.MkdirAll(filepath.Dir(link), 0755)

	if err := CreateRelativeSymlink(link, target); err != nil {
		t.Fatal(err)
	}

	got, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("..", "repo", ".vimrc")
	if got != want {
		t.Errorf("readlink = %q, want %q", got, want)
	}

	// Verify symlink resolves
	data, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "\" vim" {
		t.Errorf("content = %q, want %q", string(data), "\" vim")
	}
}
