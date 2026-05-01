package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"

	"agent-note-tracker/internal/ant"
)

const usage = `ant — Agent Notebook Tool

Usage:
  ant [--db <path>] <command> [flags]

Commands:
  init        Initialise the .ant/ directory and database
  config      Show prefix, schema version, and DB path
  add         Capture a new entry
  show        Show full detail for an entry
  edit        Update an existing entry
  delete      Permanently delete an entry
  list        List entries
  recent      Show the most recent entries
  search      Search entries by query
  for         Show entries linked to an issue id
  export      Export entries as markdown or JSON
  version     Print the build version
  completion  Print a shell completion script

Use 'ant <command> --help' for command-specific help.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ant: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	dbPath, args, err := extractDBFlag(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "-h", "--help", "help":
		fmt.Print(usage)
		return nil
	}

	app := ant.New(dbPath, stdinReader(), os.Stdout, os.Stderr)
	defer app.Close()
	return app.Dispatch(cmd, rest)
}

// stdinReader returns os.Stdin when something is piped in, and an empty
// reader when stdin is a terminal — keeping commands that fall back to
// stdin (like 'ant add') from blocking on an interactive prompt.
func stdinReader() io.Reader {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		return strings.NewReader("")
	}
	return os.Stdin
}

// extractDBFlag pulls the global --db flag (or --db=value form) from the
// front of args. It must appear before the subcommand; once the first
// non-flag argument is seen, scanning stops.
func extractDBFlag(args []string) (string, []string, error) {
	dbPath := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--db":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--db requires a value")
			}
			dbPath = args[i+1]
			args = append(args[:i], args[i+2:]...)
			i--
		case strings.HasPrefix(a, "--db="):
			dbPath = strings.TrimPrefix(a, "--db=")
			args = append(args[:i], args[i+1:]...)
			i--
		default:
			return dbPath, args, nil
		}
	}
	return dbPath, args, nil
}
