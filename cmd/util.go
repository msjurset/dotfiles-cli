package cmd

import (
	"path/filepath"
	"strings"
)

// isInside returns true if child is located within parent (resolving symlinks),
// along with the relative path of child within parent.
func isInside(parent, child string) (bool, string) {
	realParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		realParent = filepath.Clean(parent)
	}
	realChild, err := filepath.EvalSymlinks(child)
	if err != nil {
		realChild = filepath.Clean(child)
	}
	rel, err := filepath.Rel(realParent, realChild)
	if err != nil || strings.HasPrefix(rel, "..") {
		return false, ""
	}
	return true, rel
}
