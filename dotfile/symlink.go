package dotfile

import (
	"os"
	"path/filepath"
	"strings"
)

func IsManaged(path, repoPath string) (bool, string) {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		return false, ""
	}
	target, err := os.Readlink(path)
	if err != nil {
		return false, ""
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return false, ""
	}
	rel, err := filepath.Rel(repoPath, target)
	if err != nil {
		return false, ""
	}
	if strings.HasPrefix(rel, "..") {
		return false, ""
	}
	return true, rel
}

func CreateRelativeSymlink(linkPath, targetPath string) error {
	rel, err := filepath.Rel(filepath.Dir(linkPath), targetPath)
	if err != nil {
		return err
	}
	return os.Symlink(rel, linkPath)
}
