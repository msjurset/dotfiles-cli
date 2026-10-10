package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDefinitions(t *testing.T) {
	repoDir := t.TempDir()

	// Create test dotfiles
	aliasesContent := `
alias ll='ls -la'
alias caf="caffeinate -is"
`
	os.WriteFile(filepath.Join(repoDir, ".aliases"), []byte(aliasesContent), 0644)

	zshrcContent := `
export EDITOR="nvim"
function myprompt() {
    echo "prompt"
}
otherfunc() {
    echo "other"
}
`
	os.WriteFile(filepath.Join(repoDir, ".zshrc"), []byte(zshrcContent), 0644)

	// Create zfuncs file
	zfuncsDir := filepath.Join(repoDir, ".zfuncs")
	os.MkdirAll(zfuncsDir, 0755)
	os.WriteFile(filepath.Join(zfuncsDir, "mybrewapps"), []byte("#!/bin/zsh\n"), 0644)

	// Test alias match
	matches, err := FindDefinitions(repoDir, "caf")
	if err != nil {
		t.Fatalf("FindDefinitions failed: %v", err)
	}
	if len(matches) != 1 || matches[0].Kind != "alias" {
		t.Fatalf("expected 1 alias match for caf, got: %+v", matches)
	}

	// Test export match
	matches, err = FindDefinitions(repoDir, "EDITOR")
	if err != nil {
		t.Fatalf("FindDefinitions failed: %v", err)
	}
	if len(matches) != 1 || matches[0].Kind != "variable / export" {
		t.Fatalf("expected 1 export match for EDITOR, got: %+v", matches)
	}

	// Test function matches
	matches, err = FindDefinitions(repoDir, "myprompt")
	if err != nil {
		t.Fatalf("FindDefinitions failed: %v", err)
	}
	if len(matches) != 1 || matches[0].Kind != "function" {
		t.Fatalf("expected 1 function match for myprompt, got: %+v", matches)
	}

	// Test zfuncs autoload file match
	matches, err = FindDefinitions(repoDir, "mybrewapps")
	if err != nil {
		t.Fatalf("FindDefinitions failed: %v", err)
	}
	if len(matches) != 1 || matches[0].Kind != "autoload function file" {
		t.Fatalf("expected 1 autoload function file match, got: %+v", matches)
	}

	// Test nonexistent
	matches, err = FindDefinitions(repoDir, "nonexistent")
	if err != nil {
		t.Fatalf("FindDefinitions failed: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got: %+v", matches)
	}
}
