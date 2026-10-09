package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/dotfile"
	"github.com/msjurset/dotfiles/repo"
)

type IssueType string

const (
	IssueMissingSymlink     IssueType = "MISSING"
	IssueClobberedSymlink   IssueType = "CLOBBERED"
	IssueMisdirectedSymlink IssueType = "MISDIRECTED"
	IssueDanglingSymlink    IssueType = "DANGLING"
	IssueInsecurePerms      IssueType = "PERMISSIONS"
)

type DoctorIssue struct {
	Type        IssueType
	Description string
	HomePath    string
	RepoPath    string
	FixFunc     func() error
}

func Doctor(fix bool) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fmt.Println(color.Bold("Checking dotfiles health..."))
	var issues []DoctorIssue

	ignorePatterns, _ := dotfile.LoadIgnorePatterns(r.Path)

	// 1. Check repo files vs home symlinks
	walkErr := filepath.Walk(r.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(r.Path, path)
		if err != nil {
			return nil
		}

		// Only check dotfiles (top-level must start with .)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		topLevel := parts[0]
		if !strings.HasPrefix(topLevel, ".") {
			return nil
		}

		// Skip ignored files (e.g. .dotfilesignore, .dotfilesguard, .DS_Store)
		base := info.Name()
		if dotfile.IsIgnored(base, ignorePatterns) || dotfile.IsIgnored(topLevel, ignorePatterns) {
			return nil
		}

		homeTarget := filepath.Join(home, rel)

		// If an ancestor directory in home is already a managed symlink to repo,
		// the file is already managed and doesn't need its own symlink.
		if isUnderManagedDir(homeTarget, home, r.Path) {
			return nil
		}

		homeInfo, lstatErr := os.Lstat(homeTarget)
		if os.IsNotExist(lstatErr) {
			targetCopy := path
			homeCopy := homeTarget
			issues = append(issues, DoctorIssue{
				Type:        IssueMissingSymlink,
				Description: fmt.Sprintf("%s is missing symlink in home directory", color.Bold(rel)),
				HomePath:    homeTarget,
				RepoPath:    path,
				FixFunc: func() error {
					_ = os.MkdirAll(filepath.Dir(homeCopy), 0755)
					return dotfile.CreateRelativeSymlink(homeCopy, targetCopy)
				},
			})
			return nil
		}

		if lstatErr == nil {
			if homeInfo.Mode()&os.ModeSymlink == 0 {
				// Clobbered symlink: regular file exists in home where symlink should be
				targetCopy := path
				homeCopy := homeTarget
				issues = append(issues, DoctorIssue{
					Type:        IssueClobberedSymlink,
					Description: fmt.Sprintf("%s in ~ is a regular file, not a symlink to repo", color.Bold(rel)),
					HomePath:    homeTarget,
					RepoPath:    path,
					FixFunc: func() error {
						homeData, err1 := os.ReadFile(homeCopy)
						repoData, err2 := os.ReadFile(targetCopy)
						if err1 == nil && err2 == nil && !bytes.Equal(homeData, repoData) {
							// Backup original before replacing
							backup := fmt.Sprintf("%s.backup.%d", homeCopy, time.Now().Unix())
							_ = os.Rename(homeCopy, backup)
							fmt.Printf("    Backed up existing file to %s\n", color.Dim(filepath.Base(backup)))
						} else {
							_ = os.Remove(homeCopy)
						}
						return dotfile.CreateRelativeSymlink(homeCopy, targetCopy)
					},
				})
			} else {
				// It is a symlink, verify it points into the repo
				managed, linkRel := dotfile.IsManaged(homeTarget, r.Path)
				if !managed {
					targetCopy := path
					homeCopy := homeTarget
					issues = append(issues, DoctorIssue{
						Type:        IssueMisdirectedSymlink,
						Description: fmt.Sprintf("%s points outside dotfiles repo", color.Bold(rel)),
						HomePath:    homeTarget,
						RepoPath:    path,
						FixFunc: func() error {
							_ = os.Remove(homeCopy)
							return dotfile.CreateRelativeSymlink(homeCopy, targetCopy)
						},
					})
				} else if linkRel != rel {
					targetCopy := path
					homeCopy := homeTarget
					issues = append(issues, DoctorIssue{
						Type:        IssueMisdirectedSymlink,
						Description: fmt.Sprintf("%s points to %s in repo instead of %s", color.Bold(rel), linkRel, rel),
						HomePath:    homeTarget,
						RepoPath:    path,
						FixFunc: func() error {
							_ = os.Remove(homeCopy)
							return dotfile.CreateRelativeSymlink(homeCopy, targetCopy)
						},
					})
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("walking dotfiles repo: %w", walkErr)
	}

	// 2. Check permissions on sensitive directories if they exist
	sshDir := filepath.Join(home, ".ssh")
	if info, err := os.Stat(sshDir); err == nil && info.IsDir() {
		if info.Mode().Perm() > 0700 {
			dirCopy := sshDir
			issues = append(issues, DoctorIssue{
				Type:        IssueInsecurePerms,
				Description: fmt.Sprintf("~/.ssh directory has loose permissions: %04o (should be 0700)", info.Mode().Perm()),
				HomePath:    sshDir,
				FixFunc: func() error {
					return os.Chmod(dirCopy, 0700)
				},
			})
		}
	}

	// 3. Print results
	if len(issues) == 0 {
		fmt.Printf("  %s %s\n", color.Green("✓"), "All managed dotfiles and symlinks are healthy.")
		return nil
	}

	for _, issue := range issues {
		tagColor := color.Yellow
		if issue.Type == IssueClobberedSymlink || issue.Type == IssueInsecurePerms {
			tagColor = color.Red
		}
		fmt.Printf("  [%s] %s\n", tagColor(string(issue.Type)), issue.Description)
	}

	if !fix {
		fmt.Printf("\n%s detected %d issue(s). Run '%s' to automatically repair them.\n",
			color.Yellow("doctor:"),
			len(issues),
			color.Bold("dotfiles doctor --fix"))
		return nil
	}

	// Perform repairs
	fmt.Printf("\n%s\n", color.Bold("Applying fixes..."))
	fixedCount := 0
	for _, issue := range issues {
		if issue.FixFunc != nil {
			if err := issue.FixFunc(); err != nil {
				fmt.Printf("  %s fixing %s: %v\n", color.Red("failed"), issue.HomePath, err)
			} else {
				fixedCount++
				fmt.Printf("  %s %s\n", color.Green("fixed"), issue.HomePath)
			}
		}
	}

	fmt.Printf("\n%s Repaired %d of %d issue(s).\n", color.Green("Complete:"), fixedCount, len(issues))
	return nil
}

// isUnderManagedDir returns true if any ancestor directory between home and path is a symlink pointing into repoPath.
func isUnderManagedDir(path, home, repoPath string) bool {
	dir := filepath.Dir(path)
	for {
		if dir == home || dir == "." || dir == "/" || dir == "" {
			break
		}
		if info, err := os.Lstat(dir); err == nil && info.Mode()&os.ModeSymlink != 0 {
			if managed, _ := dotfile.IsManaged(dir, repoPath); managed {
				return true
			}
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	return false
}
