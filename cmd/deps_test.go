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
if command -v git >/dev/null; then
  echo "git is here"
fi
alias bogus='some_completely_nonexistent_cli_tool_12345'
`
	os.WriteFile(filepath.Join(repoDir, ".aliases"), []byte(content), 0644)

	installed, missing, err := ScanDependencies(repoDir)
	if err != nil {
		t.Fatalf("ScanDependencies failed: %v", err)
	}

	// git should be installed on any test system
	if !installed["git"] {
		t.Errorf("expected git to be reported installed")
	}

	// bogus tool should definitely be in missing
	if len(missing["some_completely_nonexistent_cli_tool_12345"]) == 0 {
		t.Errorf("expected some_completely_nonexistent_cli_tool_12345 to be reported missing")
	}
}
