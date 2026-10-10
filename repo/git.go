package repo

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

type CommitInfo struct {
	Hash    string
	Author  string
	RelDate string
	Subject string
}

func (r *Repo) Log(count int, path string) ([]CommitInfo, error) {
	if count <= 0 {
		count = 10
	}
	args := []string{"log", fmt.Sprintf("-n%d", count), "--format=%h|%an|%ar|%s"}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := r.Git(args...)
	if err != nil {
		if strings.Contains(err.Error(), "does not have any commits yet") || strings.Contains(err.Error(), "unknown revision") {
			return nil, nil
		}
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var commits []CommitInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Author:  parts[1],
				RelDate: parts[2],
				Subject: parts[3],
			})
		}
	}
	return commits, nil
}

type SnapshotInfo struct {
	Name    string
	Tag     string
	Date    string
	Subject string
}

func (r *Repo) TagExists(tag string) bool {
	_, err := r.Git("rev-parse", "--verify", "refs/tags/"+tag)
	return err == nil
}

func (r *Repo) CreateSnapshot(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = time.Now().Format("20060102-150405")
	}
	name = strings.TrimPrefix(name, "snapshot/")
	tagName := "snapshot/" + name

	if r.TagExists(tagName) {
		return "", fmt.Errorf("snapshot %q already exists", name)
	}

	// If there are uncommitted changes, commit them before tagging
	status, err := r.StatusPorcelain()
	if err == nil && strings.TrimSpace(status) != "" {
		_ = r.Add(".")
		_ = r.Commit(fmt.Sprintf("snapshot: %s", name))
	}

	_, err = r.Git("tag", "-a", tagName, "-m", fmt.Sprintf("Snapshot: %s", name))
	if err != nil {
		return "", fmt.Errorf("creating snapshot tag: %w", err)
	}
	return name, nil
}

func (r *Repo) ListSnapshots() ([]SnapshotInfo, error) {
	out, err := r.Git("tag", "-l", "--format=%(refname:short)|%(creatordate:relative)|%(authordate:relative)|%(subject)", "snapshot/*")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var list []SnapshotInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		tag := parts[0]
		name := strings.TrimPrefix(tag, "snapshot/")
		date := ""
		if len(parts) > 1 && parts[1] != "" {
			date = parts[1]
		} else if len(parts) > 2 && parts[2] != "" {
			date = parts[2]
		}
		subj := ""
		if len(parts) > 3 {
			subj = parts[3]
		}
		list = append(list, SnapshotInfo{
			Name:    name,
			Tag:     tag,
			Date:    date,
			Subject: subj,
		})
	}
	return list, nil
}

func (r *Repo) DeleteSnapshot(name string) error {
	name = strings.TrimPrefix(name, "snapshot/")
	tagName := "snapshot/" + name
	if !r.TagExists(tagName) {
		return fmt.Errorf("snapshot %q not found", name)
	}
	_, err := r.Git("tag", "-d", tagName)
	return err
}

