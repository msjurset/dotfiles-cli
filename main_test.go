package main

import (
	"os"
	"testing"
)

func TestRunVersion(t *testing.T) {
	os.Args = []string{"dotfiles", "-version"}
	err := run()
	if err != nil {
		t.Fatalf("run() with -version returned error: %v", err)
	}
}

func TestRunHelp(t *testing.T) {
	os.Args = []string{"dotfiles", "-help"}
	err := run()
	if err != nil {
		t.Fatalf("run() with -help returned error: %v", err)
	}
}

func TestRunCompletionZsh(t *testing.T) {
	os.Args = []string{"dotfiles", "-completion", "zsh"}
	err := run()
	if err != nil {
		t.Fatalf("run() with -completion zsh returned error: %v", err)
	}
}

func TestRunCompletionBash(t *testing.T) {
	os.Args = []string{"dotfiles", "-completion", "bash"}
	err := run()
	if err != nil {
		t.Fatalf("run() with -completion bash returned error: %v", err)
	}
}

func TestRunCompletionMissingShell(t *testing.T) {
	os.Args = []string{"dotfiles", "-completion"}
	err := run()
	if err == nil {
		t.Fatal("expected error for -completion without shell arg")
	}
}

func TestRunCompletionUnsupportedShell(t *testing.T) {
	os.Args = []string{"dotfiles", "-completion", "fish"}
	err := run()
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

func TestRunUnknownCommand(t *testing.T) {
	os.Args = []string{"dotfiles", "bogus"}
	err := run()
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
