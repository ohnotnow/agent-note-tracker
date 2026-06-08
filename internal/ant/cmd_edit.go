package ant

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Edit handles 'ant edit <id>'.
//
// Update semantics:
//   --body <s>    set body to literal <s>
//   --body @<p>   set body to contents of file <p>
//   --body -      set body from stdin (explicit)
//   (stdin)       if --body is not given and stdin has content, set body
//   --title <s>   set title (empty string clears)
//   --kind <s>    set kind (rejects empty)
//   --issue <s>   set issue_id (empty string clears)
//
// Flags that aren't passed leave the column unchanged. With no flags at all
// (and no piped stdin) the call is a read-only no-op that returns the entry.
//
// To grow an entry rather than replace it, see 'ant append'.
func (a *App) Edit(args []string) error {
	args, err := canonicaliseAliases(args)
	if err != nil {
		return err
	}
	id, flagArgs, err := extractPositional(args, editStringFlags)
	if err != nil {
		return NewError(CodeUsage, "%v: usage: ant edit <id> [flags]", err)
	}

	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		bodyArg string
		title   string
		kind    string
		issue   string
		visual  bool
		long    bool
	)
	fs.StringVar(&bodyArg, "body", "", "new body, @<path> to read from a file, or - for stdin")
	fs.StringVar(&title, "title", "", "new title (empty clears)")
	fs.StringVar(&kind, "kind", "", "new kind")
	fs.StringVar(&issue, "issue", "", "new issue id (empty clears)")
	fs.BoolVar(&visual, "visual", false, "open $EDITOR with the current body")
	fs.BoolVar(&long, "long", false, "return the full record (default is slim)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant edit <id> [--body <text>|@<file>|-] [--title <s>] [--kind <k>] [--issue <id>] [--visual] [--long]")
		fmt.Fprintln(a.Stderr, "")
		fmt.Fprintln(a.Stderr, "To grow an entry rather than replace it, see 'ant append'.")
		fmt.Fprintln(a.Stderr, "")
		fmt.Fprintln(a.Stderr, "Heredoc example:")
		fmt.Fprintln(a.Stderr, "  ant edit demo-AbCdE --body - <<'EOF'")
		fmt.Fprintln(a.Stderr, "  Rewritten body. Embeds 'apostrophes' without escaping.")
		fmt.Fprintln(a.Stderr, "  EOF")
		fs.PrintDefaults()
	}
	if err := a.parseFlags(fs, flagArgs); err != nil {
		return err
	}
	if id == "" {
		return NewError(CodeUsage, "usage: ant edit <id> [flags]")
	}
	if fs.NArg() != 0 {
		return NewError(CodeUsage, "usage: ant edit <id> [flags]")
	}

	set := setFlags(fs)
	if visual && set["body"] {
		return NewError(CodeUsage, "--visual and --body are mutually exclusive")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}

	var update EntryUpdate

	switch {
	case visual:
		current, err := store.GetEntry(id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return NewError(CodeNotFound, "no entry with id %q", id)
			}
			return err
		}
		edited, err := openInEditor(current.Body)
		if err != nil {
			return err
		}
		if edited == "" {
			return NewError(CodeValidationError, "editor produced an empty body; refusing to save")
		}
		if edited != current.Body {
			update.Body = &edited
		}
	default:
		body, bodyChanged, berr := a.resolveEditBody(bodyArg, set["body"])
		if berr != nil {
			return berr
		}
		if bodyChanged {
			update.Body = &body
		}
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

	entry, err := store.UpdateEntry(id, update)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError(CodeNotFound, "no entry with id %q", id)
		}
		return err
	}
	if long {
		return a.writeJSON(entry)
	}
	return a.writeJSON(entry.Slim())
}

// resolveEditBody picks the new body for an edit, returning whether body
// should be updated. Precedence:
//   1. --body literal or --body @file
//   2. piped stdin with non-empty content
//   3. otherwise, leave unchanged
func (a *App) resolveEditBody(bodyArg string, bodyFlagSet bool) (string, bool, error) {
	if bodyFlagSet {
		switch {
		case bodyArg == "-":
			body, err := a.readStdinBody()
			if err != nil {
				return "", false, err
			}
			return body, true, nil
		case strings.HasPrefix(bodyArg, "@"):
			path := bodyArg[1:]
			content, err := os.ReadFile(path)
			if err != nil {
				return "", false, fmt.Errorf("read body file %q: %w", path, err)
			}
			return strings.TrimRight(string(content), "\n\r\t "), true, nil
		}
		return strings.TrimRight(bodyArg, "\n\r\t "), true, nil
	}
	body, err := a.readStdinBody()
	if err != nil {
		return "", false, err
	}
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

// openInEditor writes initial to a temp file, runs $EDITOR (or vi), and
// returns the resulting content with trailing whitespace trimmed. The
// editor inherits the real terminal's stdin/stdout/stderr — this is the
// one place in ant where we deliberately bypass App.Stdin/Stdout, because
// interactive editors need a real tty.
func openInEditor(initial string) (string, error) {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "vi"
	}

	f, err := os.CreateTemp("", "ant-edit-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := f.WriteString(initial); err != nil {
		f.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	parts := strings.Fields(editor)
	cmdArgs := append(parts[1:], path)
	cmd := exec.Command(parts[0], cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor %q exited with error: %w", editor, err)
	}

	out, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read edited body: %w", err)
	}
	return strings.TrimRight(string(out), "\n\r\t "), nil
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
