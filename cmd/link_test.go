package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/msjurset/dotfiles/repo"
)

func TestLinkEntriesCreatesSymlinks(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	r := &repo.Repo{Path: repoDir}

	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("# zsh"), 0644)

	// linkEntries maps repo-relative paths to ~/path
	entries := []string{".zshrc"}

	// Override HOME for this test
	t.Setenv("HOME", homeDir)

	if err := linkEntries(r, entries, true); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(homeDir, ".zshrc")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink, got regular file")
	}

	data, err := os.ReadFile(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# zsh" {
		t.Errorf("content = %q, want %q", string(data), "# zsh")
	}
}

func TestLinkEntriesSkipsExisting(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	r := &repo.Repo{Path: repoDir}

	repoFile := filepath.Join(repoDir, ".zshrc")
	os.WriteFile(repoFile, []byte("# zsh"), 0644)

	// Create correct symlink already
	linkPath := filepath.Join(homeDir, ".zshrc")
	rel, _ := filepath.Rel(homeDir, repoFile)
	os.Symlink(rel, linkPath)

	t.Setenv("HOME", homeDir)

	if err := linkEntries(r, []string{".zshrc"}, true); err != nil {
		t.Fatal(err)
	}
}

func TestLinkEntriesConflict(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	r := &repo.Repo{Path: repoDir}

	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("# repo"), 0644)
	os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte("# local"), 0644)

	t.Setenv("HOME", homeDir)

	err := linkEntries(r, []string{".zshrc"}, true)
	if err == nil {
		t.Fatal("expected error for conflicting file")
	}
}

func TestLinkEntriesCreatesParentDirs(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()
	r := &repo.Repo{Path: repoDir}

	nested := filepath.Join(repoDir, ".config", "nvim", "init.vim")
	os.MkdirAll(filepath.Dir(nested), 0755)
	os.WriteFile(nested, []byte("set nocp"), 0644)

	t.Setenv("HOME", homeDir)

	if err := linkEntries(r, []string{".config/nvim/init.vim"}, true); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(homeDir, ".config", "nvim", "init.vim")
	data, err := os.ReadFile(linkPath)
	if err != nil {
		t.Fatalf("nested symlink not created: %v", err)
	}
	if string(data) != "set nocp" {
		t.Errorf("content = %q, want %q", string(data), "set nocp")
	}
}

func TestWalkRepo(t *testing.T) {
	repoDir := t.TempDir()
	r := &repo.Repo{Path: repoDir}

	// Create repo structure
	os.MkdirAll(filepath.Join(repoDir, ".git"), 0755)
	os.WriteFile(filepath.Join(repoDir, ".git", "config"), []byte(""), 0644)
	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte("# zsh"), 0644)
	os.MkdirAll(filepath.Join(repoDir, ".config", "nvim"), 0755)
	os.WriteFile(filepath.Join(repoDir, ".config", "nvim", "init.vim"), []byte("set nocp"), 0644)
	os.WriteFile(filepath.Join(repoDir, "LICENSE"), []byte("MIT"), 0644)
	os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# dotfiles"), 0644)

	entries, err := walkRepo(r)
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]bool)
	for _, e := range entries {
		found[e] = true
	}

	// Should include dotfiles
	if !found[".zshrc"] {
		t.Error(".zshrc not found")
	}
	if !found[filepath.Join(".config", "nvim", "init.vim")] {
		t.Error(".config/nvim/init.vim not found")
	}

	// Should skip .git
	for e := range found {
		if filepath.Base(e) == "config" && contains(e, ".git") {
			t.Errorf(".git contents should be skipped, found %s", e)
		}
	}

	// Should skip repo metadata
	if found["LICENSE"] {
		t.Error("LICENSE should be skipped")
	}
	if found["README.md"] {
		t.Error("README.md should be skipped")
	}
}
