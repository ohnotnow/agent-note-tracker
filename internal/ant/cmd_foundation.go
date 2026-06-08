package ant

import (
	"flag"
	"fmt"
)

// Foundation handles 'ant foundation'. Prints the project's single foundation
// entry as full JSON (same shape as 'ant show'). The full body is included
// deliberately — this is the one read path where truncation would defeat the
// goal: the foundation entry exists to anchor an agent's judgement calls at
// session start, and that only works if it can be read in full in one call.
//
// If no foundation has been recorded, exits non-zero with a friendly hint.
//
// If somehow more than one foundation exists (which the singleton check in
// InsertEntry should prevent — only reachable via direct DB manipulation),
// the most recent one is returned and a warning is written to stderr. The
// return shape stays consistent with the happy path so callers don't have to
// special-case corruption.
func (a *App) Foundation(args []string) error {
	fs := flag.NewFlagSet("foundation", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant foundation")
		fmt.Fprintln(a.Stderr, "  Print the single foundation entry, if one has been recorded.")
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return NewError(CodeUsage, "usage: ant foundation")
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}

	entries, err := store.ListEntries(ListFilter{Kind: KindFoundation})
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return NewError(CodeNotFound, "no foundation recorded — capture one with: ant add --kind foundation --body @file.md")
	}
	if len(entries) > 1 {
		fmt.Fprintf(a.Stderr,
			"warning: %d foundation entries found (singleton expected); returning the most recent\n",
			len(entries))
	}
	return a.writeJSON(entries[0])
}
