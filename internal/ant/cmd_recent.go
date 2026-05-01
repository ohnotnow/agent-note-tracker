package ant

import (
	"flag"
	"fmt"
)

// Recent handles 'ant recent [--limit N]'. Returns the N most recent entries
// (default 5) as a slim JSON array with a SnippetLen-rune body snippet so the
// agent can scan results without fetching every body in full.
func (a *App) Recent(args []string) error {
	fs := flag.NewFlagSet("recent", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	limit := 5
	fs.IntVar(&limit, "limit", limit, "maximum number of entries to return")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if limit <= 0 {
		return fmt.Errorf("--limit must be > 0")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entries, err := store.ListEntries(ListFilter{Limit: limit})
	if err != nil {
		return err
	}

	out := make([]EntryWithSnippet, len(entries))
	for i, e := range entries {
		out[i] = e.WithSnippet(SnippetLen)
	}
	return a.writeJSON(out)
}
