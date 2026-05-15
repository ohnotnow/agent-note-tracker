package ant

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

// appendSeparator is the divider stitched between the existing body and the
// appended content. A blank-line-flanked `---` renders as an <hr> in markdown,
// which is what most viewers (glow, GitHub, ant export piped to a renderer)
// turn into a visible "next bit" break.
const appendSeparator = "\n\n---\n\n"

// Append handles 'ant append <id>'.
//
// Content sources (same precedence as 'ant add'):
//
//	--body <s>    literal text
//	--body @<p>   read from file p
//	--body -      explicit stdin
//	(stdin)       implicit stdin when --body is not given
//
// The new content is joined to the existing body with a blank line, an `---`
// horizontal rule, and another blank line. Empty content is rejected — an
// append that adds nothing is almost certainly a mistake (the body would gain
// only a trailing separator).
func (a *App) Append(args []string) error {
	id, flagArgs, err := extractPositional(args, appendStringFlags)
	if err != nil {
		return NewError(CodeValidationError, "%v: usage: ant append <id> [flags]", err)
	}

	fs := flag.NewFlagSet("append", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		bodyArg string
		long    bool
	)
	fs.StringVar(&bodyArg, "body", "", "appended content, @<path> to read from a file, or - for stdin")
	fs.BoolVar(&long, "long", false, "return the full record (default is slim)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant append <id> (--body <text> | --body @<file> | --body - | stdin) [--long]")
		fmt.Fprintln(a.Stderr, "")
		fmt.Fprintln(a.Stderr, "Joins the new content onto the existing body, separated by a markdown `---` rule.")
		fmt.Fprintln(a.Stderr, "")
		fmt.Fprintln(a.Stderr, "Heredoc example:")
		fmt.Fprintln(a.Stderr, "  ant append demo-AbCdE --body - <<'EOF'")
		fmt.Fprintln(a.Stderr, "  2026-05-15 update: linked to demo-FgHiJ")
		fmt.Fprintln(a.Stderr, "  EOF")
		fs.PrintDefaults()
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if id == "" {
		return NewError(CodeValidationError, "usage: ant append <id> [flags]")
	}
	if fs.NArg() != 0 {
		return NewError(CodeValidationError, "usage: ant append <id> [flags]")
	}

	addition, err := a.resolveBody(bodyArg)
	if err != nil {
		return err
	}
	if addition == "" {
		return NewError(CodeValidationError, "nothing to append (use --body, --body @file, --body -, or pipe via stdin)")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}

	current, err := store.GetEntry(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError(CodeNotFound, "no entry with id %q", id)
		}
		return err
	}

	newBody := joinAppend(current.Body, addition)
	entry, err := store.UpdateEntry(id, EntryUpdate{Body: &newBody})
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

// joinAppend stitches addition onto existing with appendSeparator. Trailing
// whitespace on the existing body is trimmed so the separator sits cleanly,
// and an empty existing body skips the divider entirely (no leading rule).
func joinAppend(existing, addition string) string {
	trimmed := strings.TrimRight(existing, "\n\r\t ")
	if trimmed == "" {
		return addition
	}
	return trimmed + appendSeparator + addition
}

var appendStringFlags = map[string]bool{
	"body": true,
}
