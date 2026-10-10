package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDependencies(t *testing.T) {
	repoDir := t.TempDir()

	content := `
alias cat='bat --paging=never'
alias top='btm'
alias caf='echo "hello"; caffeinate -is' # inline comment
alias chained='ls; git status'
eval "$(/usr/bin/curl -s https://example.com)"
if command -v git >/dev/null; then
  echo "git is here"
fi
if command -v nonexistent_optional_tool >/dev/null; then
  echo "optional fallback"
fi
local commands=(
  "nvim : files : file type preview"
)
alias bogus='some_completely_nonexistent_cli_tool_12345' # comment here
`
	os.WriteFile(filepath.Join(repoDir, ".aliases"), []byte(content), 0644)

	installed, missing, err := ScanDependencies(repoDir)
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// git and curl should be installed on macOS/Linux test systems
	if !installed["git"] {
		t.Errorf("expected git to be reported installed")
	}
	if !installed["curl"] {
		t.Errorf("expected curl to be reported installed from eval")
	}

	// Optional probes like command -v nonexistent_optional_tool must NOT be in missing
	if len(missing["nonexistent_optional_tool"]) > 0 {
		t.Errorf("optional probe tool should not be reported missing")
	}

	// Descriptive words like "preview" from "file type preview" must NOT be in missing
	if len(missing["preview"]) > 0 {
		t.Errorf("word 'preview' from string should not be reported missing")
	}

	// bogus tool should definitely be in missing
	if len(missing["some_completely_nonexistent_cli_tool_12345"]) == 0 {
		t.Errorf("expected some_completely_nonexistent_cli_tool_12345 to be reported missing")
	}
}
