package ant

import (
	"flag"
	"fmt"
)

// For handles 'ant for <issue-id>'. Returns the slim projection of every
// entry whose issue_id exactly matches the argument.
//
// Partial matching is intentionally NOT supported in v1: issue_ids are
// opaque to ant, and prefix-matching could conflate an epic with one of its
// child issues. If you need broader matching, consider 'ant search <query>'.
func (a *App) For(args []string) error {
	fs := flag.NewFlagSet("for", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant for <issue-id>")
		fmt.Fprintln(a.Stderr, "  Returns entries whose issue_id exactly matches the argument.")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return NewError(CodeValidationError, "usage: ant for <issue-id>")
	}
	issue := fs.Arg(0)

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}
	entries, err := store.ListEntries(ListFilter{Issue: issue})
	if err != nil {
		return err
	}

	out := make([]EntrySlim, len(entries))
	for i, e := range entries {
		out[i] = e.Slim()
	}
	return a.writeEntries(out)
}
