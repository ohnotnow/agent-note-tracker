package ant

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// InitResult is the JSON payload emitted by 'ant init'.
type InitResult struct {
	DB               string `json:"db"`
	Prefix           string `json:"prefix"`
	SchemaVersion    int    `json:"schema_version"`
	Created          bool   `json:"created"`
	Rekeyed          bool   `json:"rekeyed,omitempty"`
	GitignoreUpdated bool   `json:"gitignore_updated,omitempty"`
	// Note surfaces a judgement call the user would otherwise only discover
	// later (e.g. no .git directory, so nothing was added to .gitignore).
	Note string `json:"note,omitempty"`
}

// Init handles 'ant init [--prefix <name>]'.
//
// It ensures the .ant/ directory exists, opens (or creates) the SQLite
// database, runs migrations, and writes the prefix row. The prefix is chosen
// in this order:
//  1. --prefix flag (sanitised)
//  2. existing prefix already stored in project_config
//  3. inferred from the project root basename
//
// Re-keying of existing entries when --prefix changes is implemented in
// ant-UkLWZ.2.4; for now this command updates the prefix value but does not
// touch the entries table.
func (a *App) Init(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var explicitPrefix string
	fs.StringVar(&explicitPrefix, "prefix", "", "set the issue-id prefix")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant init [--prefix <name>]")
		fs.PrintDefaults()
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}

	dbPath, err := a.DBPath()
	if err != nil {
		return err
	}

	if dbPath != MemoryDB {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return fmt.Errorf("create db dir: %w", err)
		}
	}

	store, err := a.Store()
	if err != nil {
		return err
	}

	existing, hadExisting, err := store.GetPrefix()
	if err != nil {
		return fmt.Errorf("read existing prefix: %w", err)
	}

	prefix, rekey, err := chooseInitPrefix(explicitPrefix, existing, hadExisting)
	if err != nil {
		return err
	}

	if rekey {
		if err := store.Rekey(prefix); err != nil {
			return fmt.Errorf("re-key entries: %w", err)
		}
	} else {
		if err := store.SetPrefix(prefix); err != nil {
			return fmt.Errorf("write prefix: %w", err)
		}
	}

	schema, err := store.SchemaVersion()
	if err != nil {
		return err
	}

	gitignoreUpdated := false
	noGitRepo := false
	if dbPath != MemoryDB {
		if root, rerr := FindRoot(); rerr == nil {
			updated, noGit, gerr := EnsureGitignore(root)
			if gerr != nil {
				fmt.Fprintf(a.Stderr, "warning: could not update .gitignore: %v\n", gerr)
			} else {
				gitignoreUpdated = updated
				noGitRepo = noGit
			}
		}
	}

	result := InitResult{
		DB:               dbPath,
		Prefix:           prefix,
		SchemaVersion:    schema,
		Created:          !hadExisting,
		Rekeyed:          rekey,
		GitignoreUpdated: gitignoreUpdated,
	}
	if noGitRepo {
		result.Note = "no .git directory — not adding .ant/ to .gitignore"
	}
	return a.writeJSON(result)
}

// chooseInitPrefix applies the precedence rules from Init's docstring and
// reports whether the resulting prefix differs from the existing one (the
// signal that re-keying needs to happen — see ant-UkLWZ.2.4).
func chooseInitPrefix(explicit, existing string, hadExisting bool) (string, bool, error) {
	switch {
	case explicit != "":
		s := SanitisePrefix(explicit)
		if s == "" {
			return "", false, NewError(CodeValidationError, "prefix %q has no usable characters", explicit)
		}
		return s, hadExisting && s != existing, nil
	case hadExisting:
		return existing, false, nil
	default:
		inferred, err := InferPrefix()
		if err != nil {
			return "", false, err
		}
		return inferred, false, nil
	}
}
