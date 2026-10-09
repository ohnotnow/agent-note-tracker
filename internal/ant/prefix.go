package ant

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrNoPrefix is returned when project_config has not been initialised. It
// usually means 'ant init' has not been run against this database.
var ErrNoPrefix = errors.New("project_config has no prefix row")

// GetPrefix returns the configured prefix. The bool reports whether a row
// existed in project_config; an empty string with ok=false means the
// database has been migrated but never initialised with a prefix.
func (s *Store) GetPrefix() (string, bool, error) {
	var prefix string
	err := s.DB.QueryRow(`SELECT prefix FROM project_config WHERE id = 1`).Scan(&prefix)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", false, nil
	case err != nil:
		return "", false, err
	}
	return prefix, true, nil
}

// SetPrefix writes (or replaces) the singleton prefix row. The prefix is
// stored verbatim — callers are responsible for sanitising input.
func (s *Store) SetPrefix(prefix string) error {
	_, err := s.DB.Exec(
		`INSERT INTO project_config (id, prefix) VALUES (1, ?)
			 ON CONFLICT(id) DO UPDATE SET prefix = excluded.prefix`,
		prefix,
	)
	return err
}

// Rekey atomically updates the project prefix and rewrites every entry's
// public_id using the new prefix. The integer primary keys are untouched, so
// the sqid suffix on each entry stays the same; only the prefix changes.
// Ids of this notebook's entries cited in bodies are rewritten too, so
// mentions survive; other ids with the old prefix (ait issues sharing it, say)
// are left alone.
func (s *Store) Rekey(newPrefix string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var oldPrefix string
	err = tx.QueryRow(`SELECT prefix FROM project_config WHERE id = 1`).Scan(&oldPrefix)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read prefix: %w", err)
	}

	if _, err := tx.Exec(
		`INSERT INTO project_config (id, prefix) VALUES (1, ?)
			 ON CONFLICT(id) DO UPDATE SET prefix = excluded.prefix`,
		newPrefix,
	); err != nil {
		return fmt.Errorf("update prefix: %w", err)
	}

	type row struct {
		id       int64
		publicID string
		body     string
	}
	rows, err := tx.Query(`SELECT id, public_id, body FROM entries`)
	if err != nil {
		return fmt.Errorf("load entries: %w", err)
	}
	var entries []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.publicID, &r.body); err != nil {
			rows.Close()
			return err
		}
		entries = append(entries, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	renames := make(map[string]string, len(entries))
	for _, r := range entries {
		newPublicID, err := PublicID(newPrefix, r.id)
		if err != nil {
			return err
		}
		renames[r.publicID] = newPublicID
	}

	upd, err := tx.Prepare(`UPDATE entries SET public_id = ?, body = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer upd.Close()

	for _, r := range entries {
		body := r.body
		if oldPrefix != "" {
			body = rewriteMentions(oldPrefix, body, renames)
		}
		if _, err := upd.Exec(renames[r.publicID], body, r.id); err != nil {
			return fmt.Errorf("rekey id %d: %w", r.id, err)
		}
	}
	return tx.Commit()
}
