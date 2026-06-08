package main

import (
	"errors"
	"flag"
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
  foundation  Show the project's foundation entry (singleton)
  export      Export entries as markdown or JSON
  version     Print the build version
  completion  Print a shell completion script

Use 'ant <command> --help' for command-specific help.
`

func main() {
	err := run(os.Args[1:])
	if err == nil {
		return
	}
	// '--help' on a subcommand returns flag.ErrHelp after the FlagSet has
	// already printed its usage block. Treat it as a successful, requested
	// help dump rather than a CLI error.
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	// Commands that need a specific exit code (e.g. 'self-update --check'
	// returning 1 when an update is available) wrap their result with
	// ant.ExitWithCode.
	if code, ok := ant.ExitCode(err); ok {
		os.Exit(code)
	}
	// Anything that surfaced from app.Dispatch has already been written to
	// stderr as a JSON envelope; main just sets the failing exit code.
	// Usage-class failures exit 64 (EX_USAGE), matching ait; everything else 1.
	os.Exit(ant.ExitCodeFor(err))
}

func run(args []string) error {
	dbPath, args, err := extractDBFlag(args)
	if err != nil {
		// extractDBFlag runs before Dispatch, so its errors haven't been
		// JSON-wrapped yet. Construct an App just to use WriteError —
		// keeps the on-stderr contract uniform across all error paths.
		ant.New("", nil, os.Stdout, os.Stderr).WriteError(err)
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
	case "-v", "--version":
		cmd = "version"
	}

	app := ant.New(dbPath, stdinReader(cmd), os.Stdout, os.Stderr)
	defer app.Close()
	return app.Dispatch(cmd, rest)
}

// stdinReader returns os.Stdin when something is piped in, and an empty
// reader when stdin is a terminal — keeping commands that fall back to
// stdin (like 'ant add') from blocking on an interactive prompt.
//
// 'self-update' is the exception: its confirmation prompt explicitly wants
// real stdin so the user can type y/N. Wiring it through the cmd name keeps
// every other command's terminal-stdin shielding intact.
func stdinReader(cmd string) io.Reader {
	if cmd == "self-update" {
		return os.Stdin
	}
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
				return "", nil, ant.NewError(ant.CodeUsage, "--db requires a value")
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
