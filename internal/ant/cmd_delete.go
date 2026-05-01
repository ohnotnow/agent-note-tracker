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
		return fmt.Errorf("%w: usage: ant delete <id>", err)
	}
	if id == "" {
		return fmt.Errorf("usage: ant delete <id>")
	}
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var force bool
	fs.BoolVar(&force, "force", false, "permanently delete the entry")
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: ant delete <id>")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entry, err := store.GetEntry(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("no entry with id %q", id)
		}
		return err
	}

	if !force {
		title := entry.Title
		if title == "" {
			title = "(no title)"
		}
		fmt.Fprintf(a.Stderr,
			"would delete %s — %q (%s, %s)\n"+
				"this is irreversible; refusing to act without confirmation.\n",
			entry.PublicID, title, entry.Kind, entry.CreatedAt)
		return fmt.Errorf("delete refused")
	}

	if err := store.DeleteEntry(id); err != nil {
		return err
	}
	return a.writeJSON(entry)
}
