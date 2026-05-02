package ant

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Add handles 'ant add'.
//
// Body source priority:
//  1. --body <string>      — literal body text
//  2. --body @<path>       — read from a file
//  3. stdin                — when --body is not given
//
// Other flags: --title, --kind (default 'note'), --issue.
func (a *App) Add(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		bodyArg string
		title   string
		kind    string
		issue   string
	)
	fs.StringVar(&bodyArg, "body", "", "entry body, or @<path> to read from a file")
	fs.StringVar(&title, "title", "", "optional title")
	fs.StringVar(&kind, "kind", KindNote, "entry kind (note|adr|pivot|foundation|...)")
	fs.StringVar(&issue, "issue", "", "linked issue id (free-form)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant add (--body <text> | --body @<file> | stdin) [--title <s>] [--kind <k>] [--issue <id>]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	body, err := a.resolveBody(bodyArg)
	if err != nil {
		return err
	}
	if body == "" {
		return fmt.Errorf("body is empty (use --body, --body @file, or pipe via stdin)")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	prefix, _, err := store.GetPrefix()
	if err != nil {
		return err
	}

	entry, err := store.InsertEntry(prefix, NewEntry{
		Kind:    kind,
		Title:   title,
		Body:    body,
		IssueID: issue,
	})
	if err != nil {
		return err
	}
	return a.writeJSON(entry)
}

// resolveBody applies the body-source precedence rules described on Add.
// Trailing whitespace is trimmed from any source so a piped command like
// `echo foo | ant add` does not pick up a stray newline.
func (a *App) resolveBody(bodyArg string) (string, error) {
	switch {
	case strings.HasPrefix(bodyArg, "@"):
		path := bodyArg[1:]
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read body file %q: %w", path, err)
		}
		return strings.TrimRight(string(content), "\n\r\t "), nil
	case bodyArg != "":
		return strings.TrimRight(bodyArg, "\n\r\t "), nil
	default:
		content, err := io.ReadAll(a.Stdin)
		if err != nil {
			return "", fmt.Errorf("read body from stdin: %w", err)
		}
		return strings.TrimRight(string(content), "\n\r\t "), nil
	}
}
