package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

type WhichMatch struct {
	FilePath string
	LineNum  int
	Kind     string // "alias", "function", "export", "file"
	LineText string
}

func Which(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("usage: dotfiles which <name>")
	}

	r, err := repo.Discover()
	if err != nil {
		return err
	}

	matches, err := FindDefinitions(r.Path, name)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Printf("No definition found for %s in managed dotfiles.\n", color.Bold(name))
		return nil
	}

	fmt.Printf("%s %s:\n\n", color.Bold("Found in"), color.Cyan(name))
	for _, m := range matches {
		fmt.Printf("  %s:%d  %s\n",
			color.Bold(m.FilePath),
			m.LineNum,
			color.Dim("("+m.Kind+")"))
		fmt.Printf("    %s\n\n", strings.TrimSpace(m.LineText))
	}

	return nil
}

// FindDefinitions searches files in repoPath for aliases, functions, and exports matching query.
func FindDefinitions(repoPath, query string) ([]WhichMatch, error) {
	q := strings.TrimSpace(query)
	var matches []WhichMatch

	aliasRegex := regexp.MustCompile(`(?i)^\s*alias\s+([A-Za-z0-9_.-]+)\s*=`)
	funcRegex1 := regexp.MustCompile(`(?i)^\s*function\s+([A-Za-z0-9_.-]+)`)
	funcRegex2 := regexp.MustCompile(`(?i)^\s*([A-Za-z0-9_.-]+)\s*\(\)\s*\{`)
	exportRegex := regexp.MustCompile(`(?i)^\s*(?:export\s+)?([A-Za-z0-9_]+)\s*=`)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil
		}

		// Only check dotfiles
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if !strings.HasPrefix(parts[0], ".") {
			return nil
		}

		base := info.Name()
		if base == ".dotfilesignore" || base == ".dotfilesguard" || strings.HasSuffix(base, ".backup") {
			return nil
		}

		// Check if file itself is a standalone shell function (common in .zfuncs/)
		if strings.EqualFold(base, q) && strings.Contains(rel, ".zfuncs") {
			matches = append(matches, WhichMatch{
				FilePath: rel,
				LineNum:  1,
				Kind:     "autoload function file",
				LineText: fmt.Sprintf("(autoloaded file %s)", rel),
			})
		}

		// Read file line by line
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if sm := aliasRegex.FindStringSubmatch(line); len(sm) > 1 && strings.EqualFold(sm[1], q) {
				matches = append(matches, WhichMatch{
					FilePath: rel,
					LineNum:  lineNum,
					Kind:     "alias",
					LineText: line,
				})
				continue
			}

			if sm := funcRegex1.FindStringSubmatch(line); len(sm) > 1 && strings.EqualFold(sm[1], q) {
				matches = append(matches, WhichMatch{
					FilePath: rel,
					LineNum:  lineNum,
					Kind:     "function",
					LineText: line,
				})
				continue
			}

			if sm := funcRegex2.FindStringSubmatch(line); len(sm) > 1 && strings.EqualFold(sm[1], q) {
				matches = append(matches, WhichMatch{
					FilePath: rel,
					LineNum:  lineNum,
					Kind:     "function",
					LineText: line,
				})
				continue
			}

			if sm := exportRegex.FindStringSubmatch(line); len(sm) > 1 && strings.EqualFold(sm[1], q) {
				matches = append(matches, WhichMatch{
					FilePath: rel,
					LineNum:  lineNum,
					Kind:     "variable / export",
					LineText: line,
				})
				continue
			}
		}
		return nil
	})

	return matches, err
}
