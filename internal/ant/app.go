package ant

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/fs"
	"os"
)

// App is the central handler. It owns the lazily-opened Store and the I/O
// streams commands read from and emit through. Construction is cheap and
// never fails — path resolution and database opening happen on first use, so
// commands that don't need a database (like 'version') don't pay for it.
type App struct {
	rawDBPath string
	resolved  string
	store     *Store

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// New constructs an App. dbPath is whatever the user passed via --db (may be
// empty, in which case the path is resolved on first DB use). A nil stdin is
// treated as an empty reader.
func New(dbPath string, stdin io.Reader, stdout, stderr io.Writer) *App {
	if stdin == nil {
		stdin = emptyReader{}
	}
	return &App{
		rawDBPath: dbPath,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
	}
}

type emptyReader struct{}

func (emptyReader) Read(_ []byte) (int, error) { return 0, io.EOF }

// Close releases the database handle if one was opened.
func (a *App) Close() error {
	if a.store == nil {
		return nil
	}
	err := a.store.Close()
	a.store = nil
	return err
}

// DBPath returns the resolved database path, computing and caching it on
// first call.
func (a *App) DBPath() (string, error) {
	if a.resolved != "" {
		return a.resolved, nil
	}
	p, err := ResolveDBPath(a.rawDBPath)
	if err != nil {
		return "", err
	}
	a.resolved = p
	return p, nil
}

// Store returns the open database, opening (and migrating) it if needed.
func (a *App) Store() (*Store, error) {
	if a.store != nil {
		return a.store, nil
	}
	p, err := a.DBPath()
	if err != nil {
		return nil, err
	}
	s, err := OpenStore(p)
	if err != nil {
		return nil, err
	}
	if err := s.Migrate(); err != nil {
		_ = s.Close()
		return nil, err
	}
	a.store = s
	return s, nil
}

// Dispatch routes a parsed command name to its handler. On failure, the
// error is also serialised to stderr as a {"error": {"code", "message"}}
// envelope so agents can branch on a stable code instead of string-matching
// on stderr text. flag.ErrHelp is suppressed because the FlagSet has already
// printed its usage block — that's a successful help dump, not a failure.
func (a *App) Dispatch(cmd string, args []string) error {
	err := a.dispatch(cmd, args)
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		a.WriteError(err)
	}
	return err
}

func (a *App) dispatch(cmd string, args []string) error {
	switch cmd {
	case "init":
		return a.Init(args)
	case "config":
		return a.Config(args)
	case "add":
		return a.Add(args)
	case "show":
		return a.Show(args)
	case "edit":
		return a.Edit(args)
	case "delete":
		return a.Delete(args)
	case "list":
		return a.List(args)
	case "recent":
		return a.Recent(args)
	case "search":
		return a.Search(args)
	case "for":
		return a.For(args)
	case "foundation":
		return a.Foundation(args)
	case "export":
		return a.Export(args)
	case "version":
		return a.Version(args)
	case "completion":
		return a.Completion(args)
	default:
		return NewError(CodeValidationError, "unknown command %q", cmd)
	}
}

// writeJSON pretty-prints v as JSON to stdout.
func (a *App) writeJSON(v any) error {
	enc := json.NewEncoder(a.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// writeEntries wraps a slice in {"entries": [...]} and emits as JSON. The
// envelope leaves room to add sibling fields (total counts, cursors) later
// without breaking consumers — exactly the trap a bare array form falls
// into. Caller is responsible for ensuring entries is a non-nil slice so
// the JSON is `[]` rather than `null`.
func (a *App) writeEntries(entries any) error {
	return a.writeJSON(struct {
		Entries any `json:"entries"`
	}{Entries: entries})
}

// requireInitialised opens the database and confirms 'ant init' has been
// run. For real files, returns a friendly error if the file does not exist
// rather than silently creating one. Use from any command that should refuse
// to act against an uninitialised database.
func (a *App) requireInitialised() (*Store, string, error) {
	dbPath, err := a.DBPath()
	if err != nil {
		return nil, "", err
	}
	if dbPath != MemoryDB {
		if _, statErr := os.Stat(dbPath); errors.Is(statErr, fs.ErrNotExist) {
			return nil, dbPath, NewError(CodeUninitialised, "no ant database at %s — run 'ant init' first", dbPath)
		} else if statErr != nil {
			return nil, dbPath, statErr
		}
	}
	s, err := a.Store()
	if err != nil {
		return nil, dbPath, err
	}
	_, ok, err := s.GetPrefix()
	if err != nil {
		return nil, dbPath, err
	}
	if !ok {
		return nil, dbPath, NewError(CodeUninitialised, "ant database at %s has no prefix — run 'ant init' first", dbPath)
	}
	return s, dbPath, nil
}

