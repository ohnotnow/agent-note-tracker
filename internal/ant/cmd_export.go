package ant

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

// exportFlagsWithValue is the allowlist of value-taking flags so
// extractPositional can step over them when finding the optional id.
var exportFlagsWithValue = map[string]bool{
	"kind":  true,
	"issue": true,
	"since": true,
}

// Export handles 'ant export [<id>] [filters] [--json]'.
//
// Two forms:
//   ant export <id>              one entry, in markdown
//   ant export [filters]         every matching entry, in markdown
//
// Filters: --kind, --issue, --since (same parsing as list).
//
// --json switches the output to a JSON array of full Entry records (the same
// shape as 'list --long'), suitable for piping to jq or other tools.
func (a *App) Export(args []string) error {
	id, flagArgs, err := extractPositional(args, exportFlagsWithValue)
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		kind    string
		issue   string
		sinceIn string
		asJSON  bool
	)
	fs.StringVar(&kind, "kind", "", "filter by kind")
	fs.StringVar(&issue, "issue", "", "filter by issue id")
	fs.StringVar(&sinceIn, "since", "", "created_at >= this date (YYYY-MM-DD or RFC3339)")
	fs.BoolVar(&asJSON, "json", false, "emit a JSON array instead of markdown")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant export [<id>] [--kind <k>] [--issue <id>] [--since <date>] [--json]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if id != "" && (kind != "" || issue != "" || sinceIn != "") {
		return NewError(CodeValidationError, "can't combine <id> argument with --kind/--issue/--since")
	}

	since, err := parseSince(sinceIn)
	if err != nil {
		return err
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}

	var entries []Entry
	if id != "" {
		e, err := store.GetEntry(id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return NewError(CodeNotFound, "no entry with id %q", id)
			}
			return err
		}
		entries = []Entry{e}
	} else {
		entries, err = store.ListEntries(ListFilter{
			Kind:  kind,
			Issue: issue,
			Since: since,
		})
		if err != nil {
			return err
		}
	}

	if asJSON {
		if entries == nil {
			entries = []Entry{}
		}
		return a.writeEntries(entries)
	}

	_, err = io.WriteString(a.Stdout, RenderMarkdown(entries))
	return err
}
