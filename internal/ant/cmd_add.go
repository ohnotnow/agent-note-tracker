package ant

import (
	"errors"
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
//  3. --body -             — explicit stdin
//  4. stdin                — when --body is not given
//  5. --human              — write the body in $EDITOR (pass --title alongside)
//
// Other flags: --title, --kind (default 'note'), --issue.
func (a *App) Add(args []string) error {
	args, err := canonicaliseAliases(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		bodyArg string
		title   string
		kind    string
		issue   string
		human   bool
		long    bool
		sup     string
		reason  string
	)
	fs.StringVar(&bodyArg, "body", "", "entry body, @<path> to read from a file, or - for stdin")
	fs.BoolVar(&human, "human", false, "open $EDITOR to write the body")
	fs.StringVar(&title, "title", "", "optional title")
	fs.StringVar(&kind, "kind", KindNote, "entry kind (note|adr|pivot|foundation|handover|...)")
	fs.StringVar(&issue, "issue", "", "linked issue id (free-form)")
	fs.BoolVar(&long, "long", false, "return the full record (default is slim)")
	fs.StringVar(&sup, "supersedes", "", "id of an earlier entry this one overturns, in whole or part")
	fs.StringVar(&reason, "supersedes-reason", "", "what was overturned, when it was only part (needs --supersedes)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant add (--body <text> | --body @<file> | --body - | --human | stdin) [--title <s>] [--kind <k>] [--issue <id>] [--supersedes <id> [--supersedes-reason <s>]] [--long]")
		fmt.Fprintln(a.Stderr, "")
		fmt.Fprintln(a.Stderr, "Heredoc example:")
		fmt.Fprintln(a.Stderr, "  ant add --kind adr --title \"Use sqlite\" --body - <<'EOF'")
		fmt.Fprintln(a.Stderr, "  Rationale: pure-Go driver keeps cross-compilation clean.")
		fmt.Fprintln(a.Stderr, "  EOF")
		fs.PrintDefaults()
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}

	if human && setFlags(fs)["body"] {
		return NewError(CodeUsage, "--human and --body are mutually exclusive")
	}
	if reason != "" && sup == "" {
		return NewError(CodeUsage, "--supersedes-reason needs --supersedes <id>")
	}

	var body string
	if human {
		body, err = openInEditor("")
	} else {
		body, err = a.resolveBody(bodyArg)
	}
	if err != nil {
		return err
	}
	if body == "" {
		return NewError(CodeValidationError, "body is empty (use --body, --body @file, --human, or pipe via stdin)")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	prefix, _, err := store.GetPrefix()
	if err != nil {
		return err
	}
	// Check the target before inserting, so a bad id doesn't leave an
	// entry behind without its link.
	if sup != "" {
		if _, err := store.GetEntry(sup); err != nil {
			return supersedesError(err, sup)
		}
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
	if sup != "" {
		if err := store.SetSupersedes(entry.PublicID, sup, reason); err != nil {
			return supersedesError(err, sup)
		}
	}
	if long {
		return a.writeJSON(entry)
	}
	return a.writeJSON(entry.Slim())
}

// supersedesError turns a missing --supersedes target into a clean not_found.
func supersedesError(err error, target string) error {
	if errors.Is(err, ErrNotFound) {
		return NewError(CodeNotFound, "no entry with id %q to supersede", target)
	}
	return err
}

// resolveBody applies the body-source precedence rules described on Add.
// Trailing whitespace is trimmed from any source so a piped command like
// `echo foo | ant add` does not pick up a stray newline. The literal "-" is
// treated as explicit stdin (parallel to common Unix tool convention) — this
// makes the heredoc form `ant add --body - <<EOF` parse obviously rather than
// relying on the implicit "no --body, read stdin" rule.
func (a *App) resolveBody(bodyArg string) (string, error) {
	switch {
	case bodyArg == "-":
		return a.readStdinBody()
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
		return a.readStdinBody()
	}
}

func (a *App) readStdinBody() (string, error) {
	content, err := io.ReadAll(a.Stdin)
	if err != nil {
		return "", fmt.Errorf("read body from stdin: %w", err)
	}
	return strings.TrimRight(string(content), "\n\r\t "), nil
}
