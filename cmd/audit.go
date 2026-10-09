package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

func Audit(paths []string) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var filesToScan []string
	if len(paths) > 0 {
		for _, p := range paths {
			abs := p
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(cwd, abs)
			}
			filesToScan = append(filesToScan, abs)
		}
	} else {
		// Scan all files in the dotfiles repo
		err := filepath.Walk(r.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if info.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			filesToScan = append(filesToScan, path)
			return nil
		})
		if err != nil {
			return fmt.Errorf("walking repo: %w", err)
		}
	}

	totalFindings := 0
	for _, f := range filesToScan {
		findings, err := dotfile.ScanFile(f)
		if err != nil {
			continue
		}
		if len(findings) > 0 {
			rel, err := filepath.Rel(r.Path, f)
			if err != nil {
				rel = f
			}
			fmt.Printf("%s %s:\n", color.Red("WARNING"), color.Bold(rel))
			for _, finding := range findings {
				totalFindings++
				fmt.Printf("  line %d: %s (%s)\n",
					finding.LineNumber,
					color.Yellow(finding.RuleName),
					color.Dim(finding.MatchMasked))
			}
			fmt.Println()
		}
	}

	if totalFindings == 0 {
		fmt.Println(color.Green("Audit clean: no secrets detected in scanned files."))
		return nil
	}

	return fmt.Errorf("audit detected %d potential credential(s)", totalFindings)
}

// CheckPathsForSecrets scans specific repo-relative or absolute files and prints warnings if found.
func CheckPathsForSecrets(repoPath string, paths []string, force bool) error {
	var totalFindings []dotfile.SecretFinding
	for _, p := range paths {
		target := p
		if !filepath.IsAbs(target) {
			target = filepath.Join(repoPath, target)
		}
		findings, err := dotfile.ScanFile(target)
		if err == nil && len(findings) > 0 {
			totalFindings = append(totalFindings, findings...)
		}
	}

	if len(totalFindings) == 0 {
		return nil
	}

	if !force {
		fmt.Printf("\n%s Potential credentials detected in files to be committed:\n", color.Red("ERROR:"))
		for _, f := range totalFindings {
			rel, _ := filepath.Rel(repoPath, f.FilePath)
			if rel == "" {
				rel = f.FilePath
			}
			fmt.Printf("  %s:%d: %s (%s)\n",
				color.Bold(rel),
				f.LineNumber,
				color.Yellow(f.RuleName),
				color.Dim(f.MatchMasked))
		}
		return fmt.Errorf("secret scanner blocked commit (use --force or -f to override)")
	}

	fmt.Printf("\n%s Proceeding with commit despite %d detected secret warning(s) (--force used):\n",
		color.Yellow("WARNING:"), len(totalFindings))
	for _, f := range totalFindings {
		rel, _ := filepath.Rel(repoPath, f.FilePath)
		if rel == "" {
			rel = f.FilePath
		}
		fmt.Printf("  %s:%d: %s (%s)\n",
			color.Bold(rel),
			f.LineNumber,
			color.Yellow(f.RuleName),
			color.Dim(f.MatchMasked))
	}
	return nil
}
