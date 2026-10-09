package dotfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanBytes(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedRule string
	}{
		{
			name:         "AWS Access Key",
			content:      "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			expectedRule: "AWS Access Key ID",
		},
		{
			name:         "GitHub Token",
			content:      "GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvwxyzAB",
			expectedRule: "GitHub Personal Access Token",
		},
		{
			name:         "Google API Key",
			content:      "export GEMINI_API_KEY=AIzaSyD-1234567890abcdef1234567890abc",
			expectedRule: "Google API Key",
		},
		{
			name:         "OpenAI API Key",
			content:      "OPENAI_API_KEY=sk-proj-abc1234567890abcdefghijklmnopqrstuv",
			expectedRule: "OpenAI / Anthropic API Key",
		},
		{
			name:         "Private Key Header",
			content:      "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaA==\n-----END OPENSSH PRIVATE KEY-----",
			expectedRule: "Private Key Header",
		},
		{
			name:         "Slack Token",
			content:      "SLACK_BOT_TOKEN=xoxb-1234567890-abcdefghij",
			expectedRule: "Slack Token",
		},
		{
			name:         "Clean content without secrets",
			content:      "export PATH=$PATH:$HOME/bin\nalias ll='ls -la'\n",
			expectedRule: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := ScanBytes([]byte(tt.content), "testfile")
			if tt.expectedRule == "" {
				if len(findings) > 0 {
					t.Fatalf("expected 0 findings, got %d", len(findings))
				}
				return
			}

			if len(findings) == 0 {
				t.Fatalf("expected finding for rule %q, got none", tt.expectedRule)
			}

			matched := false
			for _, f := range findings {
				if f.RuleName == tt.expectedRule {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("expected finding with rule %q, got: %+v", tt.expectedRule, findings)
			}
		})
	}
}

func TestScanFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".test_secrets")
	os.WriteFile(path, []byte("AKIAIOSFODNN7EXAMPLE\n"), 0644)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatalf("ScanFile failed: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].MatchMasked != "AKIA...MPLE" {
		t.Errorf("unexpected masked value: %s", findings[0].MatchMasked)
	}
}

func TestMaskSecret(t *testing.T) {
	if got := MaskSecret("short"); got != "********" {
		t.Errorf("MaskSecret(short) = %q, want ********", got)
	}
	if got := MaskSecret("123456789012"); got != "1234...9012" {
		t.Errorf("MaskSecret(long) = %q, want 1234...9012", got)
	}
}
