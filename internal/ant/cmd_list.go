package ant

import (
	"flag"
	"fmt"
	"text/tabwriter"
	"time"
)

// List handles 'ant list'.
//
// Default output: slim JSON array.
// Flags:
//   --long             include the full body (returns []Entry)
//   --human            tabular text for humans (date / kind / id / title)
//   --kind <k>         filter by kind
//   --issue <id>       filter by issue_id
//   --since <date>     created_at >= date (YYYY-MM-DD or RFC3339)
func (a *App) List(args []string) error {
	args, err := canonicaliseAliases(args)
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	var (
		long    bool
		human   bool
		kind    string
		issue   string
		sinceIn string
	)
	fs.BoolVar(&long, "long", false, "include the full body")
	fs.BoolVar(&human, "human", false, "tabular human-readable view")
	fs.StringVar(&kind, "kind", "", "filter by kind")
	fs.StringVar(&issue, "issue", "", "filter by issue id")
	fs.StringVar(&sinceIn, "since", "", "created_at >= this date (YYYY-MM-DD or RFC3339)")
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant list [--long | --human] [--kind <k>] [--issue <id>] [--since <date>]")
		fs.PrintDefaults()
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}
	if long && human {
		return NewError(CodeUsage, "--long and --human are mutually exclusive")
	}

	since, err := parseSince(sinceIn)
	if err != nil {
		return err
	}

	store, _, err := a.requireInitialised()
	if err != nil {
		return err
	}

	entries, err := store.ListEntries(ListFilter{
		Kind:  kind,
		Issue: issue,
		Since: since,
	})
	if err != nil {
		return err
	}

	switch {
	case human:
		return a.writeHumanTable(entries)
	case long:
		if entries == nil {
			entries = []Entry{}
		}
		return a.writeEntries(entries)
	default:
		slim := make([]EntrySlim, len(entries))
		for i, e := range entries {
			slim[i] = e.Slim()
		}
		return a.writeEntries(slim)
	}
}

// writeHumanTable renders an aligned date/kind/id/title table to stdout.
func (a *App) writeHumanTable(entries []Entry) error {
	w := tabwriter.NewWriter(a.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "DATE\tKIND\tID\tTITLE"); err != nil {
		return err
	}
	for _, e := range entries {
		title := e.Title
		if title == "" {
			title = "—"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			shortDate(e.CreatedAt), e.Kind, e.PublicID, title); err != nil {
			return err
		}
	}
	return w.Flush()
}

// shortDate truncates RFC3339 to the date portion.
func shortDate(rfc3339 string) string {
	if len(rfc3339) >= 10 {
		return rfc3339[:10]
	}
	return rfc3339
}

// parseSince accepts YYYY-MM-DD or RFC3339 and returns a UTC RFC3339 string
// suitable for direct comparison against entries.created_at.
func parseSince(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", NewError(CodeValidationError, "could not parse --since %q (try YYYY-MM-DD or RFC3339)", s)
}
