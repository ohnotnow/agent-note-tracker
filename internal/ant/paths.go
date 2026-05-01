package ant

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	DBDirName  = ".ant"
	DBFileName = "ant.db"
	MemoryDB   = ":memory:"
)

// ResolveDBPath returns the absolute path to the SQLite database.
//
// If explicit is non-empty it is returned as-is, including the special value
// ":memory:" used by tests. Otherwise the path is <project-root>/.ant/ant.db,
// where project root is the git top level (or the current working directory
// if there is no git repo).
func ResolveDBPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	root, err := FindRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, DBDirName, DBFileName), nil
}

// FindRoot returns the project root: the git top-level directory if the
// current directory is inside a git repo, otherwise the current working
// directory.
func FindRoot() (string, error) {
	if root, err := gitTopLevel(); err == nil {
		return root, nil
	}
	return os.Getwd()
}

func gitTopLevel() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// InferPrefix returns the default issue-id prefix derived from the project
// root's basename (lowercased, with non-alphanumerics stripped). Falls back
// to "ant" if the resulting string would be empty — only InferPrefix
// applies that fallback; SanitisePrefix returns the empty string so its
// callers can choose to reject unusable input instead.
func InferPrefix() (string, error) {
	root, err := FindRoot()
	if err != nil {
		return "", err
	}
	if p := SanitisePrefix(filepath.Base(root)); p != "" {
		return p, nil
	}
	return "ant", nil
}

// SanitisePrefix normalises a candidate prefix: lowercased ASCII alphanumerics
// only. Returns the empty string if no usable characters remain.
func SanitisePrefix(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		}
	}
	return b.String()
}
