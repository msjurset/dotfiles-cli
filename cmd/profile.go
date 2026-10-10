package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Profile(name string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return RunProfile(r.Path, home, name)
}

func RunProfile(repoPath, home, name string) error {
	profilesDir := filepath.Join(repoPath, "profiles")
	activeFile := filepath.Join(repoPath, ".active_profile")

	// 1. If no argument, list available profiles
	if strings.TrimSpace(name) == "" {
		return listProfiles(profilesDir, activeFile)
	}

	// 2. Clear / deactivate profile
	if name == "--clear" || name == "clear" || name == "none" {
		return clearActiveProfile(repoPath, home, profilesDir, activeFile)
	}

	// 3. Switch to profile
	targetDir := filepath.Join(profilesDir, name)
	info, err := os.Stat(targetDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("profile %q not found in %s", name, profilesDir)
	}

	// Clear previous profile if one was active
	_ = clearActiveProfile(repoPath, home, profilesDir, activeFile)

	// Link all files from profiles/<name> into $HOME
	linkedCount := 0
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(targetDir, path)
		if err != nil {
			return nil
		}
		homeTarget := filepath.Join(home, rel)
		_ = os.MkdirAll(filepath.Dir(homeTarget), 0755)

		// Create symlink to the profile file
		_ = os.Remove(homeTarget)
		if err := dotfile.CreateRelativeSymlink(homeTarget, path); err != nil {
			return fmt.Errorf("linking %s: %w", rel, err)
		}
		linkedCount++
		return nil
	})
	if err != nil {
		return err
	}

	// Save active profile name
	if err := os.WriteFile(activeFile, []byte(name+"\n"), 0644); err != nil {
		return fmt.Errorf("saving active profile: %w", err)
	}

	fmt.Printf("%s Switched to profile %s (%d file(s) linked).\n",
		color.Green("profile:"), color.Bold(name), linkedCount)
	return nil
}

func listProfiles(profilesDir, activeFile string) error {
	entries, err := os.ReadDir(profilesDir)
	if err != nil || len(entries) == 0 {
		fmt.Printf("No profiles found.\nTo create one, add a directory: %s/<name>/\n", profilesDir)
		return nil
	}

	activeProfile := ""
	if data, err := os.ReadFile(activeFile); err == nil {
		activeProfile = strings.TrimSpace(string(data))
	}

	fmt.Println(color.Bold("Available profiles:"))
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		pName := e.Name()
		if pName == activeProfile {
			fmt.Printf("  * %s %s\n", color.Green(pName), color.Bold("(active)"))
		} else {
			fmt.Printf("    %s\n", pName)
		}
	}
	return nil
}

func clearActiveProfile(repoPath, home, profilesDir, activeFile string) error {
	data, err := os.ReadFile(activeFile)
	if err != nil {
		return nil // No active profile
	}
	activeName := strings.TrimSpace(string(data))
	if activeName == "" {
		return nil
	}

	targetDir := filepath.Join(profilesDir, activeName)
	_ = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(targetDir, path)
		if err != nil {
			return nil
		}
		homeTarget := filepath.Join(home, rel)
		// Check if it is a symlink pointing to this profile file
		if managed, _ := dotfile.IsManaged(homeTarget, targetDir); managed {
			_ = os.Remove(homeTarget)
		}
		return nil
	})

	_ = os.Remove(activeFile)
	fmt.Printf("%s Deactivated profile %s.\n", color.Yellow("profile:"), color.Bold(activeName))
	return nil
}
