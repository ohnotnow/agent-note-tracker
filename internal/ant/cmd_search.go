package ant

import (
	"flag"
	"fmt"
	"strings"
)

// Search handles 'ant search <query> [more terms...]'.
//
// Multi-word queries are split on whitespace and ANDed: 'ant search auth
// refactor' matches entries containing both 'auth' AND 'refactor' somewhere
// in title or body (case-insensitive). v1 has no FTS5; LIKE is fine for
// notebooks of any reasonable size.
func (a *App) Search(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant search <query> [more terms...]")
		fmt.Fprintln(a.Stderr, "  Multi-term queries are ANDed across title and body, case-insensitive.")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	terms := strings.Fields(strings.Join(fs.Args(), " "))
	if len(terms) == 0 {
		return NewError(CodeValidationError, "usage: ant search <query>")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entries, err := store.ListEntries(ListFilter{SearchTerms: terms})
	if err != nil {
		return err
	}

	out := make([]EntryWithSnippet, len(entries))
	for i, e := range entries {
		out[i] = e.WithSnippet(SnippetLen)
	}
	return a.writeEntries(out)
}
