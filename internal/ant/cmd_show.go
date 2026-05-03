package ant

import (
	"errors"
	"flag"
	"fmt"
)

// Show handles 'ant show <id>'. Outputs the full entry record as JSON, or a
// clean error if the id is not found.
func (a *App) Show(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant show <id>")
		fmt.Fprintln(a.Stderr, "  Print the full record for one entry as JSON.")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return NewError(CodeValidationError, "usage: ant show <id>")
	}
	id := fs.Arg(0)

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
	return a.writeJSON(entry)
}
