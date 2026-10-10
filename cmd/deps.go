package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/msjurset/dotfiles/color"
	"github.com/msjurset/dotfiles/repo"
)

type ToolReference struct {
	Tool     string
	FilePath string
	LineNum  int
	Context  string
}

var shellBuiltins = map[string]bool{
	"echo": true, "cd": true, "ls": true, "cat": true, "grep": true,
	"sed": true, "awk": true, "source": true, "export": true, "alias": true,
	"local": true, "return": true, "exit": true, "test": true, "true": true,
	"false": true, "eval": true, "exec": true, "read": true, "printf": true,
	"setopt": true, "unsetopt": true, "autoload": true, "compdef": true,
	"zstyle": true, "bindkey": true, "rm": true, "cp": true, "mv": true,
	"mkdir": true, "rmdir": true, "touch": true, "chmod": true, "chown": true,
	"sleep": true, "kill": true, "wait": true, "trap": true, "umask": true,
	"pushd": true, "popd": true, "dirs": true, "fg": true, "bg": true,
	"jobs": true, "history": true, "man": true, "head": true, "tail": true,
	"sort": true, "uniq": true, "wc": true, "tr": true, "find": true,
	"xargs": true, "cut": true, "tee": true, "which": true, "type": true,
	"command": true, "builtin": true, "hash": true, "pwd": true, "env": true,
	"for": true, "while": true, "if": true, "then": true, "else": true,
	"elif": true, "fi": true, "case": true, "esac": true, "in": true,
	"do": true, "done": true, "function": true, "let": true, "set": true,
	"shift": true, "unset": true, "print": true,
	".": true, ":": true,
}

func Deps(strict bool) error {
	r, err := repo.Discover()
	if err != nil {
		return err
	}

	fmt.Println(color.Bold("Checking CLI tool dependencies referenced in dotfiles..."))
	installed, missing, err := ScanDependencies(r.Path)
	if err != nil {
		return err
	}

	if len(installed) > 0 {
		var toolNames []string
		for t := range installed {
			toolNames = append(toolNames, t)
		}
		sort.Strings(toolNames)
		fmt.Printf("  %s %s\n\n",
			color.Green("✓"),
			strings.Join(toolNames, ", "))
	}

	if len(missing) == 0 {
		fmt.Println(color.Green("All referenced CLI tools are installed!"))
		return nil
	}

	fmt.Printf("%s %d tool(s) referenced in dotfiles are not found in PATH:\n",
		color.Red("Missing:"), len(missing))

	var missingNames []string
	for tool, refs := range missing {
		missingNames = append(missingNames, tool)
		first := refs[0]
		fmt.Printf("  %s %-12s %s\n",
			color.Red("✗"),
			color.Bold(tool),
			color.Dim(fmt.Sprintf("referenced in %s:%d (%s)", first.FilePath, first.LineNum, first.Context)))
	}
	sort.Strings(missingNames)

	fmt.Println()
	fmt.Printf("Suggested install:\n  %s %s\n",
		color.Cyan("brew install"),
		strings.Join(missingNames, " "))

	if strict {
		return fmt.Errorf("%d missing tool dependency/dependencies detected", len(missing))
	}
	return nil
}

// ScanDependencies inspects shell scripts in repoPath and checks if invoked tools exist in PATH.
func ScanDependencies(repoPath string) (map[string]bool, map[string][]ToolReference, error) {
	installed := make(map[string]bool)
	missing := make(map[string][]ToolReference)

	aliasTargetRegex := regexp.MustCompile(`(?i)^\s*alias\s+[A-Za-z0-9_.-]+\s*=\s*['"]?([A-Za-z0-9_.-]+)`)
	checkCmdRegex := regexp.MustCompile(`(?i)(?:command\s+-v|type|which)\s+([A-Za-z0-9_.-]+)`)

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

		parts := strings.Split(filepath.ToSlash(rel), "/")
		if !strings.HasPrefix(parts[0], ".") {
			return nil
		}

		base := info.Name()
		if base == ".dotfilesignore" || base == ".dotfilesguard" || strings.HasSuffix(base, ".backup") {
			return nil
		}

		// Only check shell scripts and dotfiles
		ext := filepath.Ext(base)
		isShell := ext == ".sh" || ext == ".zsh" || ext == ".bash" ||
			base == ".aliases" || base == ".zshrc" || base == ".zprofile" ||
			base == ".zshenv" || base == ".bashrc" || base == ".profile" ||
			strings.Contains(rel, ".zfuncs") || strings.Contains(rel, ".zsh.d")

		if !isShell {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			var candidates []string
			if sm := aliasTargetRegex.FindStringSubmatch(line); len(sm) > 1 {
				candidates = append(candidates, sm[1])
			}
			if sm := checkCmdRegex.FindStringSubmatch(line); len(sm) > 1 {
				candidates = append(candidates, sm[1])
			}

			for _, cmdName := range candidates {
				cmdName = strings.Trim(cmdName, `"'`)
				if cmdName == "" || shellBuiltins[cmdName] || strings.HasPrefix(cmdName, "-") || strings.Contains(cmdName, "/") {
					continue
				}

				if _, err := exec.LookPath(cmdName); err == nil {
					installed[cmdName] = true
				} else if defs, _ := FindDefinitions(repoPath, cmdName); len(defs) > 0 {
					// Defined internally as a shell function, alias, or autoload file
					installed[cmdName] = true
				} else {
					missing[cmdName] = append(missing[cmdName], ToolReference{
						Tool:     cmdName,
						FilePath: rel,
						LineNum:  lineNum,
						Context:  line,
					})
				}
			}
		}
		return nil
	})

	return installed, missing, err
}
