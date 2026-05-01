package ant

import "flag"

// Version is the build version. Override at build time with:
//
//	go build -ldflags "-X agent-note-tracker/internal/ant.Version=v0.1.0"
//
// The default of "dev" makes a development build obvious in the version
// output.
var Version = "dev"

// Version handles 'ant version'. Emits a small JSON record so it composes
// with jq the same way other commands do.
//
// The GitHub release-check (compare current Version against the latest tag)
// is intentionally left out until ant-UkLWZ.8.3 picks the public repo
// location — there's nothing to check against yet.
func (a *App) Version(args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	return a.writeJSON(map[string]string{"version": Version})
}
