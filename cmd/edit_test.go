package cmd

import (
	"testing"
)

func TestMatchDotfile(t *testing.T) {
	files := []string{
		".aliases",
		".bashrc",
		".tmux.conf",
		".vimrc",
		".zshrc",
		".config/nvim/init.vim",
		".config/alacritty/alacritty.toml",
	}

	tests := []struct {
		query    string
		expected string
	}{
		{"zsh", ".zshrc"},
		{".zshrc", ".zshrc"},
		{"zshrc", ".zshrc"},
		{"alias", ".aliases"},
		{"tmux", ".tmux.conf"},
		{"vim", ".vimrc"},
		{"nvim", ".config/nvim/init.vim"},
		{"alacritty", ".config/alacritty/alacritty.toml"},
		{"nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := matchDotfile(tt.query, files)
			if got != tt.expected {
				t.Errorf("matchDotfile(%q) = %q, want %q", tt.query, got, tt.expected)
			}
		})
	}
}
