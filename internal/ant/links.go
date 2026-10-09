package ant

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// LinkSupersedes is the links.kind for "this entry overturns (part of) that
// one". It is the only declared link: lineage is left to mentions, which need
// nothing from the writer.
const LinkSupersedes = "supersedes"

// ErrNoLink is returned by RemoveSupersedes when there is no such link.
var ErrNoLink = errors.New("link not found")

// SetSupersedes records that the entry from supersedes the entry to, with an
// optional reason. Declaring an existing link again replaces its reason.
// Returns ErrNotFound if either entry is missing.
func (s *Store) SetSupersedes(from, to, reason string) error {
	if from == to {
		return fmt.Errorf("an entry cannot supersede itself")
	}
	fromEntry, err := s.GetEntry(from)
	if err != nil {
		return err
	}
	toEntry, err := s.GetEntry(to)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(
		`INSERT INTO links (from_entry, to_entry, kind, reason, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (from_entry, to_entry, kind) DO UPDATE SET reason = excluded.reason`,
		fromEntry.id, toEntry.id, LinkSupersedes, sqlNullable(reason),
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// RemoveSupersedes deletes the link saying from supersedes to. Returns
// ErrNoLink if there was none.
func (s *Store) RemoveSupersedes(from, to string) error {
	res, err := s.DB.Exec(
		`DELETE FROM links
		 WHERE kind = ?
		   AND from_entry = (SELECT id FROM entries WHERE public_id = ?)
		   AND to_entry   = (SELECT id FROM entries WHERE public_id = ?)`,
		LinkSupersedes, from, to,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoLink
	}
	return nil
}

// attachSupersededBy fills SupersededBy on each entry with the public ids of
// the entries that supersede it, oldest link first.
func (s *Store) attachSupersededBy(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	rows, err := s.DB.Query(
		`SELECT l.to_entry, e.public_id
		 FROM links l JOIN entries e ON e.id = l.from_entry
		 WHERE l.kind = ?
		 ORDER BY l.created_at, l.rowid`,
		LinkSupersedes,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	by := map[int64][]string{}
	for rows.Next() {
		var (
			to   int64
			from string
		)
		if err := rows.Scan(&to, &from); err != nil {
			return err
		}
		by[to] = append(by[to], from)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range entries {
		entries[i].SupersededBy = by[entries[i].id]
	}
	return nil
}

// supersedeRefs returns the entries that id supersedes and the entries that
// supersede it, each with title and reason, for 'ant show'.
func (s *Store) supersedeRefs(id int64) (supersedes, supersededBy []EntryRef, err error) {
	supersedes, err = s.linkRefs(
		`SELECT e.public_id, e.title, l.reason
		 FROM links l JOIN entries e ON e.id = l.to_entry
		 WHERE l.kind = ? AND l.from_entry = ?
		 ORDER BY l.created_at, l.rowid`, id)
	if err != nil {
		return nil, nil, err
	}
	supersededBy, err = s.linkRefs(
		`SELECT e.public_id, e.title, l.reason
		 FROM links l JOIN entries e ON e.id = l.from_entry
		 WHERE l.kind = ? AND l.to_entry = ?
		 ORDER BY l.created_at, l.rowid`, id)
	return supersedes, supersededBy, err
}

func (s *Store) linkRefs(query string, id int64) ([]EntryRef, error) {
	rows, err := s.DB.Query(query, LinkSupersedes, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []EntryRef
	for rows.Next() {
		var (
			ref           EntryRef
			title, reason sql.NullString
		)
		if err := rows.Scan(&ref.PublicID, &title, &reason); err != nil {
			return nil, err
		}
		ref.Title = title.String
		ref.Reason = reason.String
		out = append(out, ref)
	}
	return out, rows.Err()
}
