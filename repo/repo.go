package repo

import (
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	Path string
}

func Discover() (*Repo, error) {
	p := os.Getenv("DOTFILES_REPO_PATH")
	if p == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home dir: %w", err)
		}
		p = filepath.Join(home, ".dotfiles")
	}
	info, err := os.Stat(p)
	if err != nil {
		return nil, fmt.Errorf("dotfiles repo not found at %s: %w", p, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", p)
	}
	return &Repo{Path: p}, nil
}
