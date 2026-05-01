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
// the sqid suffix on each entry stays the same — only the prefix changes.
func (s *Store) Rekey(newPrefix string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`INSERT INTO project_config (id, prefix) VALUES (1, ?)
			 ON CONFLICT(id) DO UPDATE SET prefix = excluded.prefix`,
		newPrefix,
	); err != nil {
		return fmt.Errorf("update prefix: %w", err)
	}

	rows, err := tx.Query(`SELECT id FROM entries`)
	if err != nil {
		return fmt.Errorf("load entry ids: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	upd, err := tx.Prepare(`UPDATE entries SET public_id = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer upd.Close()

	for _, id := range ids {
		newPublicID, err := PublicID(newPrefix, id)
		if err != nil {
			return err
		}
		if _, err := upd.Exec(newPublicID, id); err != nil {
			return fmt.Errorf("rekey id %d: %w", id, err)
		}
	}
	return tx.Commit()
}
