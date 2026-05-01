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
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: ant show <id>")
	}
	id := fs.Arg(0)

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
	return a.writeJSON(entry)
}
