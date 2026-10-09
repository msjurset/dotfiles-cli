package dotfile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultGuardrails is the built-in list of sensitive credential and secret patterns.
var DefaultGuardrails = []string{
	".ssh",
	".gnupg",
	".gpg",
	"id_rsa*",
	"id_ed25519*",
	"id_ecdsa*",
	"id_dsa*",
	"*.pem",
	"*.key",
	"*.p12",
	"*.pfx",
	"*.pkcs12",
	".secrets*",
	".env*",
	"*secret*",
	"*credential*",
	"*token*",
	".aws",
	".azure",
	".netrc",
	".dockercfg",
}

// LoadGuardPatterns loads custom sensitive patterns from ~/.dotfiles/.dotfilesguard
// and combines them with DefaultGuardrails.
func LoadGuardPatterns(repoPath string) ([]string, error) {
	var patterns []string
	guardFile := filepath.Join(repoPath, ".dotfilesguard")

	if f, err := os.Open(guardFile); err == nil {
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
			return nil, fmt.Errorf("reading .dotfilesguard: %w", err)
		}
	}

	patterns = append(patterns, DefaultGuardrails...)
	return patterns, nil
}

// IsSensitive checks if a path or base name matches any sensitive guardrail pattern.
// Returns true and the matched pattern if sensitive.
func IsSensitive(path string, patterns []string) (bool, string) {
	base := filepath.Base(path)
	cleanPath := filepath.Clean(path)

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Match against base name
		if matched, err := filepath.Match(p, base); err == nil && matched {
			return true, p
		}
		// Match against full path components
		if strings.Contains(cleanPath, p) {
			return true, p
		}
	}
	return false, ""
}

// AppendGuardPattern adds new sensitive patterns to .dotfilesguard in the repo.
func AppendGuardPattern(repoPath string, patterns ...string) ([]string, error) {
	guardFile := filepath.Join(repoPath, ".dotfilesguard")

	existing := make(map[string]bool)
	if f, err := os.Open(guardFile); err == nil {
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
	f, err := os.OpenFile(guardFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening .dotfilesguard: %w", err)
	}
	defer f.Close()

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" || existing[p] {
			continue
		}
		if _, err := fmt.Fprintln(f, p); err != nil {
			return nil, fmt.Errorf("writing to .dotfilesguard: %w", err)
		}
		existing[p] = true
		added = append(added, p)
	}

	return added, nil
}
