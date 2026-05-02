package ant

import (
	"flag"
	"fmt"
)

// ConfigResult is the JSON payload emitted by 'ant config'.
type ConfigResult struct {
	DB            string `json:"db"`
	Prefix        string `json:"prefix"`
	SchemaVersion int    `json:"schema_version"`
}

// Config handles 'ant config' — reports the resolved DB path, prefix, and
// schema version. Refuses to run against an uninitialised database so that
// merely asking 'are we set up?' can't accidentally create the file.
func (a *App) Config(args []string) error {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant config")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, dbPath, err := a.requireInitialised()
	if err != nil {
		return err
	}

	prefix, _, err := store.GetPrefix()
	if err != nil {
		return err
	}
	schema, err := store.SchemaVersion()
	if err != nil {
		return err
	}
	return a.writeJSON(ConfigResult{
		DB:            dbPath,
		Prefix:        prefix,
		SchemaVersion: schema,
	})
}
