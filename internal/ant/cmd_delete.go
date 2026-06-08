package ant

import (
	"errors"
	"flag"
	"fmt"
)

// Delete handles 'ant delete <id> [--force]'.
//
// Without --force the command prints a warning to stderr describing the
// entry that would be deleted and exits non-zero, leaving the database
// untouched. With --force the row is removed and the deleted record is
// echoed to stdout as JSON so the action is visible in scrollback.
//
// The --force requirement is documented only in the README — the Claude Code
// skill deliberately omits it. See ant-UkLWZ (Q4 in the initiative
// description) for the reasoning.
func (a *App) Delete(args []string) error {
	id, flagArgs, err := extractPositional(args, nil)
	if err != nil {
		return NewError(CodeUsage, "%v: usage: ant delete <id>", err)
	}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		force bool
		long  bool
	)
	fs.BoolVar(&force, "force", false, "permanently delete the entry")
	fs.BoolVar(&long, "long", false, "echo the full record (default is slim)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant delete <id> [--force] [--long]")
		fs.PrintDefaults()
	}
	if err := a.parseFlags(fs, flagArgs); err != nil {
		return err
	}
	if id == "" {
		return NewError(CodeUsage, "usage: ant delete <id>")
	}
	if fs.NArg() != 0 {
		return NewError(CodeUsage, "usage: ant delete <id>")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entry, err := store.GetEntry(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return NewError(CodeNotFound, "no entry with id %q", id)
		}
		return err
	}

	if !force {
		title := entry.Title
		if title == "" {
			title = "(no title)"
		}
		return NewError(CodeConfirmationRequired,
			"would delete %s — %q (%s, %s); this is irreversible. Re-run with --force to proceed.",
			entry.PublicID, title, entry.Kind, entry.CreatedAt)
	}

	if err := store.DeleteEntry(id); err != nil {
		return err
	}
	if long {
		return a.writeJSON(entry)
	}
	return a.writeJSON(entry.Slim())
}
