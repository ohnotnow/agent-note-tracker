package ant

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Edit handles 'ant edit <id>'.
//
// Update semantics:
//   --body <s>    set body to literal <s>
//   --body @<p>   set body to contents of file <p>
//   (stdin)       if --body is not given and stdin has content, set body
//   --title <s>   set title (empty string clears)
//   --kind <s>    set kind (rejects empty)
//   --issue <s>   set issue_id (empty string clears)
//
// Flags that aren't passed leave the column unchanged. With no flags at all
// (and no piped stdin) the call is a read-only no-op that returns the entry.
func (a *App) Edit(args []string) error {
	id, flagArgs, err := extractPositional(args, editStringFlags)
	if err != nil {
		return fmt.Errorf("%w: usage: ant edit <id> [flags]", err)
	}
	if id == "" {
		return fmt.Errorf("usage: ant edit <id> [flags]")
	}

	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		bodyArg string
		title   string
		kind    string
		issue   string
	)
	fs.StringVar(&bodyArg, "body", "", "new body, or @<path> to read from a file")
	fs.StringVar(&title, "title", "", "new title (empty clears)")
	fs.StringVar(&kind, "kind", "", "new kind")
	fs.StringVar(&issue, "issue", "", "new issue id (empty clears)")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: ant edit <id> [flags]")
	}

	set := setFlags(fs)

	var update EntryUpdate

	body, bodyChanged, err := a.resolveEditBody(bodyArg, set["body"])
	if err != nil {
		return err
	}
	if bodyChanged {
		update.Body = &body
	}
	if set["title"] {
		update.Title = &title
	}
	if set["kind"] {
		update.Kind = &kind
	}
	if set["issue"] {
		update.IssueID = &issue
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entry, err := store.UpdateEntry(id, update)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("no entry with id %q", id)
		}
		return err
	}
	return a.writeJSON(entry)
}

// resolveEditBody picks the new body for an edit, returning whether body
// should be updated. Precedence:
//   1. --body literal or --body @file
//   2. piped stdin with non-empty content
//   3. otherwise, leave unchanged
func (a *App) resolveEditBody(bodyArg string, bodyFlagSet bool) (string, bool, error) {
	if bodyFlagSet {
		if strings.HasPrefix(bodyArg, "@") {
			path := bodyArg[1:]
			content, err := os.ReadFile(path)
			if err != nil {
				return "", false, fmt.Errorf("read body file %q: %w", path, err)
			}
			return strings.TrimRight(string(content), "\n\r\t "), true, nil
		}
		return strings.TrimRight(bodyArg, "\n\r\t "), true, nil
	}
	stdinBytes, err := io.ReadAll(a.Stdin)
	if err != nil {
		return "", false, fmt.Errorf("read body from stdin: %w", err)
	}
	body := strings.TrimRight(string(stdinBytes), "\n\r\t ")
	if body == "" {
		return "", false, nil
	}
	return body, true, nil
}

// setFlags returns the names of flags explicitly passed by the user. Use
// this to distinguish "unset" (Go's default zero value) from "explicitly
// passed empty string", which matters for clear semantics.
func setFlags(fs *flag.FlagSet) map[string]bool {
	out := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { out[f.Name] = true })
	return out
}

// editStringFlags lists the flag names that consume a value, so
// extractPositional can step over `--flag value` pairs without confusing the
// value with the positional id.
var editStringFlags = map[string]bool{
	"body":  true,
	"title": true,
	"kind":  true,
	"issue": true,
}

// extractPositional pulls at most one non-flag argument out of args and
// returns it alongside the remaining flag-shaped args. It tolerates flags
// appearing before or after the positional, which Go's stdlib flag package
// does not. Returns ("", rest, nil) when no positional is found — callers
// that require one should check for the empty string.
//
// flagsWithValue contains the flag names that consume a following arg
// (--flag value form); it is consulted to avoid mistaking a flag value for
// the positional argument. The "--flag=value" form is also handled.
func extractPositional(args []string, flagsWithValue map[string]bool) (string, []string, error) {
	var (
		positional string
		found      bool
		rest       []string
	)
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			for _, r := range args[i+1:] {
				if found {
					return "", nil, fmt.Errorf("unexpected extra argument %q", r)
				}
				positional = r
				found = true
			}
			break
		}
		if strings.HasPrefix(a, "--") {
			rest = append(rest, a)
			name := strings.TrimPrefix(a, "--")
			if strings.Contains(name, "=") {
				continue
			}
			if flagsWithValue[name] && i+1 < len(args) {
				rest = append(rest, args[i+1])
				i++
			}
			continue
		}
		if found {
			return "", nil, fmt.Errorf("unexpected extra argument %q", a)
		}
		positional = a
		found = true
	}
	return positional, rest, nil
}
