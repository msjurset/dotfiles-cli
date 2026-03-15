package dotfile

import (
	"os"
	"path/filepath"
)

type Entry struct {
	Name    string
	AbsPath string
	IsDir   bool
	Managed bool
	RepoRel string
}

func Scan(dir, repoPath string) ([]Entry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []Entry
	for _, e := range entries {
		name := e.Name()
		if name == "." || name == ".." || name[0] != '.' {
			continue
		}
		abs := filepath.Join(dir, name)
		info, err := os.Lstat(abs)
		if err != nil {
			continue
		}
		ent := Entry{
			Name:    name,
			AbsPath: abs,
			IsDir:   info.IsDir(),
		}
		managed, rel := IsManaged(abs, repoPath)
		if managed {
			ent.Managed = true
			ent.RepoRel = rel
		}
		result = append(result, ent)
	}
	return result, nil
}
