package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileManagement(t *testing.T) {
	repoDir := t.TempDir()
	homeDir := t.TempDir()

	profilesDir := filepath.Join(repoDir, "profiles")
	workDir := filepath.Join(profilesDir, "work")
	personalDir := filepath.Join(profilesDir, "personal")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(personalDir, 0755); err != nil {
		t.Fatal(err)
	}

	workFile := filepath.Join(workDir, ".gitconfig.local")
	if err := os.WriteFile(workFile, []byte("[user]\n    email = work@example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}

	personalFile := filepath.Join(personalDir, ".gitconfig.local")
	if err := os.WriteFile(personalFile, []byte("[user]\n    email = personal@example.com\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. List profiles with no active profile
	if err := RunProfile(repoDir, homeDir, ""); err != nil {
		t.Fatalf("expected no error listing profiles, got: %v", err)
	}

	// 2. Switch to 'work'
	if err := RunProfile(repoDir, homeDir, "work"); err != nil {
		t.Fatalf("expected no error switching to work profile, got: %v", err)
	}

	// Verify .active_profile has "work"
	activeData, err := os.ReadFile(filepath.Join(repoDir, ".active_profile"))
	if err != nil {
		t.Fatalf("expected .active_profile to exist, got: %v", err)
	}
	if strings.TrimSpace(string(activeData)) != "work" {
		t.Fatalf("expected active profile 'work', got: %q", string(activeData))
	}

	// Verify symlink created in home
	homeGitConfig := filepath.Join(homeDir, ".gitconfig.local")
	info, err := os.Lstat(homeGitConfig)
	if err != nil {
		t.Fatalf("expected ~/.gitconfig.local to exist, got: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected ~/.gitconfig.local to be a symlink")
	}

	content, err := os.ReadFile(homeGitConfig)
	if err != nil || !strings.Contains(string(content), "work@example.com") {
		t.Fatalf("expected work gitconfig content, got: %s", string(content))
	}

	// 3. Switch to 'personal' (should replace work symlinks with personal)
	if err := RunProfile(repoDir, homeDir, "personal"); err != nil {
		t.Fatalf("expected no error switching to personal profile, got: %v", err)
	}

	activeData, err = os.ReadFile(filepath.Join(repoDir, ".active_profile"))
	if err != nil || strings.TrimSpace(string(activeData)) != "personal" {
		t.Fatalf("expected active profile 'personal', got: %q", string(activeData))
	}

	content, err = os.ReadFile(homeGitConfig)
	if err != nil || !strings.Contains(string(content), "personal@example.com") {
		t.Fatalf("expected personal gitconfig content, got: %s", string(content))
	}

	// 4. Deactivate with --clear
	if err := RunProfile(repoDir, homeDir, "--clear"); err != nil {
		t.Fatalf("expected no error clearing profile, got: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repoDir, ".active_profile")); !os.IsNotExist(err) {
		t.Fatalf("expected .active_profile to be removed")
	}

	if _, err := os.Lstat(homeGitConfig); !os.IsNotExist(err) {
		t.Fatalf("expected ~/.gitconfig.local symlink to be removed")
	}

	// 5. Test nonexistent profile
	if err := RunProfile(repoDir, homeDir, "nonexistent"); err == nil {
		t.Fatalf("expected error switching to nonexistent profile, got nil")
	}
}
