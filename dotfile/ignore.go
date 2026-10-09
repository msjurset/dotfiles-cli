package dotfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultIgnores is the built-in list of common OS, runtime, and cache patterns.
var DefaultIgnores = []string{
	".git",
	".dotfiles",
	".dotfilesignore",
	".dotfilesguard",
	".DS_Store",
	".Trash",
	".CFUserTextEncoding",
	"*.padl",
	".InstallAnywhere",
	".*history",
	".lesshst",
	".viminfo",
	".zcompdump*",
	".node_repl_history",
	".Rhistory",
	".Rapp.history",
	".cache",
	".npm",
	".cargo",
	".rustup",
	".bun",
	".bundle",
	".gradle",
	".gem",
	".pyenv",
	".tldrc",
	".local",
	".*_sessions",
	".*.bak",
	"*.save",
	"*.backup",
	"*.db",
	"*.sqlite*",
	".*_history.json",
	".oh-my-zsh",
	".z",
	".venv*",
	".virtualenvs",
	".pytest_cache",
	".mypy_cache",
	".ruff_cache",
	".*.swp",
	".*.swo",
	"*~",
	".terraform",
	".vagrant.d",
	".serverless",
}

// LoadIgnorePatterns loads custom patterns from ~/.dotfiles/.dotfilesignore
// and combines them with DefaultIgnores.
func LoadIgnorePatterns(repoPath string) ([]string, error) {
	var patterns []string
	ignoreFile := filepath.Join(repoPath, ".dotfilesignore")

	if f, err := os.Open(ignoreFile); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading .dotfilesignore: %w", err)
		}
	}

	patterns = append(patterns, DefaultIgnores...)
	return patterns, nil
}

// IsIgnored checks if a given dotfile name matches any ignore pattern.
func IsIgnored(name string, patterns []string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if matched, err := filepath.Match(p, name); err == nil && matched {
			return true
		}
		// Also support substring match if pattern doesn't contain glob stars
		if p == name {
			return true
		}
	}
	return false
}

// AppendIgnorePattern adds new ignore patterns to .dotfilesignore in the repo.
func AppendIgnorePattern(repoPath string, patterns ...string) ([]string, error) {
	ignoreFile := filepath.Join(repoPath, ".dotfilesignore")

	existing := make(map[string]bool)
	if f, err := os.Open(ignoreFile); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				existing[line] = true
			}
		}
		f.Close()
	}

	var added []string
	f, err := os.OpenFile(ignoreFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening .dotfilesignore: %w", err)
	}
	defer f.Close()

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" || existing[p] {
			continue
		}
		if _, err := fmt.Fprintln(f, p); err != nil {
			return nil, fmt.Errorf("writing to .dotfilesignore: %w", err)
		}
		existing[p] = true
		added = append(added, p)
	}

	return added, nil
}
