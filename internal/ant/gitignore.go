package ant

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// gitignoreEntry is the line written to .gitignore. Trailing slash is
// idiomatic for directories.
const gitignoreEntry = ".ant/"

// EnsureGitignore appends gitignoreEntry to <root>/.gitignore if a git repo
// is present and the entry is not already listed. Returns whether the file
// was modified and — so Init can surface the gap in its output — whether it
// was skipped because there is no .git directory at root. Safe to call
// unconditionally from Init.
func EnsureGitignore(root string) (updated bool, noGit bool, err error) {
	if _, err := os.Stat(filepath.Join(root, ".git")); errors.Is(err, fs.ErrNotExist) {
		return false, true, nil
	} else if err != nil {
		return false, false, err
	}

	path := filepath.Join(root, ".gitignore")
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		content = nil
	} else if err != nil {
		return false, false, err
	}

	if gitignoreContains(content, ".ant") {
		return false, false, nil
	}

	var leader string
	if len(content) > 0 && content[len(content)-1] != '\n' {
		leader = "\n"
	}
	out := append(content, []byte(leader+gitignoreEntry+"\n")...)
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return false, false, err
	}
	return true, false, nil
}

// gitignoreContains reports whether the given file content already lists
// target (matched against trimmed, non-comment, non-negation lines, with a
// trailing slash or "/*" tolerated).
func gitignoreContains(content []byte, target string) bool {
	for _, raw := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		line = strings.TrimSuffix(line, "/*")
		line = strings.TrimSuffix(line, "/**")
		line = strings.TrimPrefix(line, "/")
		if line == target {
			return true
		}
	}
	return false
}
