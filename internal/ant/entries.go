package ant

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned by Store lookups when no row matches the given id.
var ErrNotFound = errors.New("entry not found")

// NewEntry are the user-supplied fields for a fresh entry. Body is required;
// title and issueID may be empty (empty values are stored as SQL NULL).
type NewEntry struct {
	Kind    string
	Title   string
	Body    string
	IssueID string
}

// InsertEntry adds a new entry and returns the canonical record. The integer
// primary key and the public_id are derived inside a single transaction so
// the public_id is guaranteed to match the row's id.
func (s *Store) InsertEntry(prefix string, e NewEntry) (Entry, error) {
	if prefix == "" {
		return Entry{}, fmt.Errorf("InsertEntry: empty prefix")
	}
	if e.Body == "" {
		return Entry{}, fmt.Errorf("InsertEntry: empty body")
	}
	if e.Kind == "" {
		e.Kind = KindNote
	}

	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.DB.Begin()
	if err != nil {
		return Entry{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if e.Kind == KindFoundation {
		var existing string
		err := tx.QueryRow(`SELECT public_id FROM entries WHERE kind = ? LIMIT 1`, KindFoundation).Scan(&existing)
		switch {
		case err == nil:
			return Entry{}, NewError(CodeConflict, "a foundation entry already exists (%s); use 'ant edit %s' to revise it", existing, existing)
		case errors.Is(err, sql.ErrNoRows):
			// fine, no existing foundation
		default:
			return Entry{}, fmt.Errorf("check existing foundation: %w", err)
		}
	}

	var nextID int64
	if err := tx.QueryRow(`SELECT COALESCE(MAX(id), 0) + 1 FROM entries`).Scan(&nextID); err != nil {
		return Entry{}, fmt.Errorf("compute next id: %w", err)
	}

	publicID, err := PublicID(prefix, nextID)
	if err != nil {
		return Entry{}, err
	}

	if _, err := tx.Exec(
		`INSERT INTO entries
			(id, public_id, kind, title, body, issue_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		nextID, publicID, e.Kind, sqlNullable(e.Title), e.Body, sqlNullable(e.IssueID), now, now,
	); err != nil {
		return Entry{}, fmt.Errorf("insert entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Entry{}, err
	}

	return Entry{
		id:        nextID,
		PublicID:  publicID,
		Kind:      e.Kind,
		Title:     e.Title,
		Body:      e.Body,
		IssueID:   e.IssueID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetEntry returns a single entry by its public_id. Returns ErrNotFound if
// no row matches.
func (s *Store) GetEntry(publicID string) (Entry, error) {
	var (
		e        Entry
		title    sql.NullString
		issueID  sql.NullString
	)
	row := s.DB.QueryRow(
		`SELECT id, public_id, kind, title, body, issue_id, created_at, updated_at
		 FROM entries WHERE public_id = ?`,
		publicID,
	)
	if err := row.Scan(&e.id, &e.PublicID, &e.Kind, &title, &e.Body, &issueID, &e.CreatedAt, &e.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, err
	}
	if title.Valid {
		e.Title = title.String
	}
	if issueID.Valid {
		e.IssueID = issueID.String
	}
	return e, nil
}

func sqlNullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// DeleteEntry removes the row with the given public_id. Returns ErrNotFound
// if no row matches.
func (s *Store) DeleteEntry(publicID string) error {
	res, err := s.DB.Exec(`DELETE FROM entries WHERE public_id = ?`, publicID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// EntryUpdate carries optional column changes for UpdateEntry. A non-nil
// pointer means "set this column" — an empty string in Title/IssueID clears
// the column to NULL. Body and Kind reject empty values.
type EntryUpdate struct {
	Body    *string
	Title   *string
	Kind    *string
	IssueID *string
}

// UpdateEntry applies the changes in u to the entry with the given public_id
// and returns the resulting record. Returns ErrNotFound if no row matches.
// A no-op update (every field nil) still returns the current record.
func (s *Store) UpdateEntry(publicID string, u EntryUpdate) (Entry, error) {
	var (
		sets []string
		argv []any
	)
	if u.Body != nil {
		if *u.Body == "" {
			return Entry{}, fmt.Errorf("body cannot be empty")
		}
		sets = append(sets, "body = ?")
		argv = append(argv, *u.Body)
	}
	if u.Title != nil {
		sets = append(sets, "title = ?")
		argv = append(argv, sqlNullable(*u.Title))
	}
	if u.Kind != nil {
		if *u.Kind == "" {
			return Entry{}, fmt.Errorf("kind cannot be empty")
		}
		sets = append(sets, "kind = ?")
		argv = append(argv, *u.Kind)
	}
	if u.IssueID != nil {
		sets = append(sets, "issue_id = ?")
		argv = append(argv, sqlNullable(*u.IssueID))
	}
	if len(sets) == 0 {
		return s.GetEntry(publicID)
	}

	sets = append(sets, "updated_at = ?")
	now := time.Now().UTC().Format(time.RFC3339)
	argv = append(argv, now, publicID)

	q := "UPDATE entries SET "
	for i, c := range sets {
		if i > 0 {
			q += ", "
		}
		q += c
	}
	q += " WHERE public_id = ?"

	res, err := s.DB.Exec(q, argv...)
	if err != nil {
		return Entry{}, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Entry{}, err
	}
	if n == 0 {
		return Entry{}, ErrNotFound
	}
	return s.GetEntry(publicID)
}

// ListFilter constrains the rows returned by ListEntries. Zero values mean
// "no filter on this field". SearchTerms applies a case-insensitive LIKE per
// term against title and body, ANDed — so ["auth", "refactor"] matches entries
// containing both words across either column. Since must already be in
// RFC3339 format.
type ListFilter struct {
	Kind        string
	Issue       string
	Since       string
	SearchTerms []string
	Limit       int
}

// ListEntries returns rows matching filter, newest first.
func (s *Store) ListEntries(f ListFilter) ([]Entry, error) {
	q := `SELECT id, public_id, kind, title, body, issue_id, created_at, updated_at
	      FROM entries`
	var (
		clauses []string
		argv    []any
	)
	if f.Kind != "" {
		clauses = append(clauses, "kind = ?")
		argv = append(argv, f.Kind)
	}
	if f.Issue != "" {
		clauses = append(clauses, "issue_id = ?")
		argv = append(argv, f.Issue)
	}
	if f.Since != "" {
		clauses = append(clauses, "created_at >= ?")
		argv = append(argv, f.Since)
	}
	for _, term := range f.SearchTerms {
		if term == "" {
			continue
		}
		clauses = append(clauses, "(LOWER(title) LIKE ? OR LOWER(body) LIKE ?)")
		needle := "%" + lower(term) + "%"
		argv = append(argv, needle, needle)
	}
	if len(clauses) > 0 {
		q += " WHERE " + joinAnd(clauses)
	}
	q += " ORDER BY created_at DESC, id DESC"
	if f.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", f.Limit)
	}

	rows, err := s.DB.Query(q, argv...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Entry
	for rows.Next() {
		var (
			e       Entry
			title   sql.NullString
			issueID sql.NullString
		)
		if err := rows.Scan(&e.id, &e.PublicID, &e.Kind, &title, &e.Body, &issueID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		if title.Valid {
			e.Title = title.String
		}
		if issueID.Valid {
			e.IssueID = issueID.String
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func joinAnd(parts []string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out += " AND " + p
	}
	return out
}
