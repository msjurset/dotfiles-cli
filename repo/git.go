package repo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func (r *Repo) Git(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", r.Path}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (r *Repo) StatusPorcelain(paths ...string) (string, error) {
	args := []string{"status", "--porcelain"}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	return r.Git(args...)
}

func (r *Repo) Add(paths ...string) error {
	args := []string{"add"}
	args = append(args, paths...)
	_, err := r.Git(args...)
	return err
}

func (r *Repo) Commit(msg string) error {
	_, err := r.Git("commit", "-m", msg)
	return err
}

func (r *Repo) Push() error {
	_, err := r.Git("push")
	return err
}

func (r *Repo) PullRebase() error {
	_, err := r.Git("pull", "--rebase")
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "no tracking information") ||
			strings.Contains(errStr, "no remote repository specified") ||
			strings.Contains(errStr, "couldn't find remote ref") ||
			strings.Contains(errStr, "There is no tracking information") {
			return nil
		}
		return err
	}
	return nil
}

func (r *Repo) Diff(staged bool, paths ...string) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}
	return r.Git(args...)
}

func (r *Repo) UnpushedCount() (int, error) {
	out, err := r.Git("rev-list", "@{u}..HEAD", "--count")
	if err != nil {
		// No upstream configured or unborn branch
		if strings.Contains(err.Error(), "no upstream") || strings.Contains(err.Error(), "unknown revision") || strings.Contains(err.Error(), "no such branch") {
			return 0, nil
		}
		return 0, err
	}
	n, err := strconv.Atoi(out)
	if err != nil {
		return 0, fmt.Errorf("parsing unpushed count %q: %w", out, err)
	}
	return n, nil
}
