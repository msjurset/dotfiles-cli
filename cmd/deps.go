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

var shellCache = make(map[string]bool)

func isShellCommand(name string) bool {
	if res, ok := shellCache[name]; ok {
		return res
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "zsh"
	}
	var cmd *exec.Cmd
	if strings.Contains(shell, "zsh") {
		cmd = exec.Command(shell, "-i", "-c", "whence -w "+name)
	} else {
		cmd = exec.Command(shell, "-i", "-c", "type -t "+name)
	}
	out, err := cmd.Output()
	res := false
	if err == nil {
		s := strings.TrimSpace(string(out))
		if strings.Contains(s, ": function") || strings.Contains(s, ": builtin") || strings.Contains(s, ": alias") || strings.Contains(s, ": command") || s == "function" || s == "builtin" || s == "alias" || s == "file" {
			res = true
		}
	}
	shellCache[name] = res
	return res
}

// ScanDependencies inspects shell scripts in repoPath and checks if invoked tools exist in PATH or shell environment.
func ScanDependencies(repoPath string) (map[string]bool, map[string][]ToolReference, error) {
	installed := make(map[string]bool)
	missing := make(map[string][]ToolReference)

	aliasRegex := regexp.MustCompile(`(?i)^\s*alias\s+[A-Za-z0-9_.-]+\s*=\s*(.+)`)
	evalCmdRegex := regexp.MustCompile(`(?i)\beval\s+["']?\$\(\s*([^\s)]+)`)
	probeCmdRegex := regexp.MustCompile(`(?i)(?:^|[;&|]\s*|\bif\s+|\belif\s+)(?:command\s+-v|which)\s+([A-Za-z0-9_.-]+)`)

	var checkCandidate = func(cmdName, rel string, lineNum int, line string, isRequired bool) {
		cmdName = strings.Trim(cmdName, `"'`)
		if strings.Contains(cmdName, "/") {
			cmdName = filepath.Base(cmdName)
		}
		if cmdName == "" || shellBuiltins[cmdName] || strings.HasPrefix(cmdName, "-") {
			return
		}
		if installed[cmdName] {
			return
		}

		if _, err := exec.LookPath(cmdName); err == nil {
			installed[cmdName] = true
			return
		}

		if defs, _ := FindDefinitions(repoPath, cmdName); len(defs) > 0 {
			installed[cmdName] = true
			return
		}

		if isShellCommand(cmdName) {
			installed[cmdName] = true
			return
		}

		// Only add to missing if it is an actual requirement, not an optional probe
		if isRequired {
			missing[cmdName] = append(missing[cmdName], ToolReference{
				Tool:     cmdName,
				FilePath: rel,
				LineNum:  lineNum,
				Context:  line,
			})
		}
	}

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

			// 1. Check alias commands (required)
			if sm := aliasRegex.FindStringSubmatch(line); len(sm) > 1 {
				val := strings.TrimSpace(sm[1])
				if strings.HasPrefix(val, "'") {
					if idx := strings.Index(val[1:], "'"); idx != -1 {
						val = val[1 : 1+idx]
					}
				} else if strings.HasPrefix(val, "\"") {
					if idx := strings.Index(val[1:], "\""); idx != -1 {
						val = val[1 : 1+idx]
					}
				} else {
					if idx := strings.Index(val, "#"); idx != -1 {
						val = val[:idx]
					}
				}

				// Split on command separators (;, &&, ||, |)
				parts := strings.FieldsFunc(val, func(r rune) bool {
					return r == ';' || r == '|' || r == '&'
				})
				for _, part := range parts {
					words := strings.Fields(strings.TrimSpace(part))
					if len(words) > 0 {
						checkCandidate(words[0], rel, lineNum, line, true)
					}
				}
			}

			// 2. Check eval "$(tool ...)" invocations (required)
			if sm := evalCmdRegex.FindStringSubmatch(line); len(sm) > 1 {
				checkCandidate(sm[1], rel, lineNum, line, true)
			}

			// 3. Check command -v / which probes (optional - recorded if installed, not reported as missing)
			if sm := probeCmdRegex.FindStringSubmatch(line); len(sm) > 1 {
				checkCandidate(sm[1], rel, lineNum, line, false)
			}
		}
		return nil
	})

	return installed, missing, err
}
