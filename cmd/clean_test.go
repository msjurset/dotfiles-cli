package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")

	os.WriteFile(src, []byte("hello"), 0644)

	if err := copyFile(src, dst); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q, want %q", string(data), "hello")
	}
}

func TestCopyDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "srcdir")
	dst := filepath.Join(dir, "dstdir")

	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("bbb"), 0644)

	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path    string
		content string
	}{
		{filepath.Join(dst, "a.txt"), "aaa"},
		{filepath.Join(dst, "sub", "b.txt"), "bbb"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			data, err := os.ReadFile(tt.path)
			if err != nil {
				t.Fatalf("file not found: %v", err)
			}
			if string(data) != tt.content {
				t.Errorf("content = %q, want %q", string(data), tt.content)
			}
		})
	}
}

func TestCopyPathFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "file")
	dst := filepath.Join(dir, "copy")
	os.WriteFile(src, []byte("data"), 0644)

	if err := copyPath(src, dst); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("content = %q, want %q", string(data), "data")
	}
}

func TestCopyPathDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "d")
	dst := filepath.Join(dir, "dcopy")
	os.MkdirAll(src, 0755)
	os.WriteFile(filepath.Join(src, "f"), []byte("x"), 0644)

	if err := copyPath(src, dst); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "f"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "x" {
		t.Errorf("content = %q, want %q", string(data), "x")
	}
}
