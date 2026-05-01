package ant

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store wraps the SQLite handle and exposes the query helpers used by the
// command handlers.
type Store struct {
	DB *sql.DB
}

// OpenStore opens the database at path. Pass ":memory:" for an in-memory
// database (used in tests).
func OpenStore(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("empty database path")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite %q: %w", path, err)
	}
	return &Store{DB: db}, nil
}

// Close closes the underlying database handle. Safe to call on a nil receiver.
func (s *Store) Close() error {
	if s == nil || s.DB == nil {
		return nil
	}
	return s.DB.Close()
}
