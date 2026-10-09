package dotfile

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type SecretFinding struct {
	RuleName    string
	LineNumber  int
	MatchMasked string
	FilePath    string
}

type secretRule struct {
	Name    string
	Pattern *regexp.Regexp
}

var secretRules = []secretRule{
	{
		Name:    "Private Key Header",
		Pattern: regexp.MustCompile(`-----BEGIN [A-Z0-9_-]+ PRIVATE KEY-----`),
	},
	{
		Name:    "AWS Access Key ID",
		Pattern: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	},
	{
		Name:    "AWS Secret Access Key",
		Pattern: regexp.MustCompile(`(?i)(?:aws_secret_access_key|aws_sec_key)\s*[:=]\s*["']?([A-Za-z0-9/+=]{40})["']?`),
	},
	{
		Name:    "GitHub Personal Access Token",
		Pattern: regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36,}\b`),
	},
	{
		Name:    "GitHub Fine-Grained PAT",
		Pattern: regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{82}\b`),
	},
	{
		Name:    "Google API Key",
		Pattern: regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{30,40}\b`),
	},
	{
		Name:    "OpenAI / Anthropic API Key",
		Pattern: regexp.MustCompile(`\bsk-(?:proj-|ant-)?[A-Za-z0-9_-]{20,}\b`),
	},
	{
		Name:    "Slack Token",
		Pattern: regexp.MustCompile(`\bxox[baprs]-[0-9A-Za-z-]{10,}\b`),
	},
}

// MaskSecret returns a masked version of a detected credential string.
func MaskSecret(secret string) string {
	if len(secret) <= 8 {
		return "********"
	}
	prefix := secret[:4]
	suffix := secret[len(secret)-4:]
	return fmt.Sprintf("%s...%s", prefix, suffix)
}

// ScanBytes scans a byte slice for secret patterns and returns all findings.
func ScanBytes(content []byte, filePath string) []SecretFinding {
	var findings []SecretFinding
	scanner := bufio.NewScanner(bytes.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			// Skip pure comment lines if desired, but check for inline secrets
			if !strings.Contains(line, "BEGIN") && !strings.Contains(line, "AKIA") &&
				!strings.Contains(line, "AIza") && !strings.Contains(line, "ghp_") &&
				!strings.Contains(line, "sk-") {
				continue
			}
		}

		for _, rule := range secretRules {
			locs := rule.Pattern.FindAllString(line, -1)
			for _, loc := range locs {
				findings = append(findings, SecretFinding{
					RuleName:    rule.Name,
					LineNumber:  lineNum,
					MatchMasked: MaskSecret(loc),
					FilePath:    filePath,
				})
			}
		}
	}
	return findings
}

// ScanFile scans an entire file on disk for secrets.
func ScanFile(path string) ([]SecretFinding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ScanBytes(data, path), nil
}
