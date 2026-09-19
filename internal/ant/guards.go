package ant

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

// Guards against blind writes by coding agents. Mirrors ait's guards.go so
// an agent hopping between the two tools meets one contract:
//
//   - The read check: an agent running under Claude Code (which sets
//     CLAUDE_CODE_SESSION_ID in the environment of every command) may only
//     rewrite an entry it has `show`n in the same session within ShownWindow.
//     Humans in a plain terminal have no session id and are never checked.
//     Override with SkipReadCheckFlag.
//   - The shrink check: a new body less than half the length of the existing
//     one is refused unless --force is given, on the theory that a diff or a
//     summary was sent by mistake. Growing a body is never questioned.

// ShownWindow is how long a `show` counts as having read the entry. Context
// compaction keeps the same session id, so a stale show must expire.
const ShownWindow = time.Hour

// SkipReadCheckFlag is deliberately long: typing it is an admission that the
// caller is editing something it has not read.
const SkipReadCheckFlag = "--dangerously-skip-read-check"

// ShrinkGuardMinLength is the existing-body length below which the shrink
// check does not apply; short bodies get rewritten wholesale all the time.
const ShrinkGuardMinLength = 200

// CodeUnread and CodeShrink are the error codes the two guards emit. Same
// spelling as ait.
const (
	CodeUnread = "unread"
	CodeShrink = "shrink"
)

const sessionEnvVar = "CLAUDE_CODE_SESSION_ID"

func currentSessionID() string {
	return strings.TrimSpace(os.Getenv(sessionEnvVar))
}

// MarkShown records that the current session has displayed the entry. A run
// with no session id records nothing.
func (s *Store) MarkShown(publicID string) error {
	session := currentSessionID()
	if session == "" {
		return nil
	}
	_, err := s.DB.Exec(
		`UPDATE entries SET shown_by_session = ?, shown_at = ? WHERE public_id = ?`,
		session, time.Now().UTC().Format(time.RFC3339), publicID)
	return err
}

// CheckShown refuses the write unless the entry was shown in this session
// within ShownWindow. skip is the value of SkipReadCheckFlag.
func (s *Store) CheckShown(publicID string, skip bool) error {
	session := currentSessionID()
	if session == "" || skip {
		return nil
	}
	var shownBy, shownAt sql.NullString
	err := s.DB.QueryRow(
		`SELECT shown_by_session, shown_at FROM entries WHERE public_id = ?`, publicID,
	).Scan(&shownBy, &shownAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var why string
	switch {
	case !shownBy.Valid || !shownAt.Valid:
		why = "it has never been shown in this session"
	case shownBy.String != session:
		why = "it was last shown by a different session"
	default:
		t, err := time.Parse(time.RFC3339, shownAt.String)
		if err != nil {
			return fmt.Errorf("parse shown_at %q: %w", shownAt.String, err)
		}
		age := time.Since(t)
		if age <= ShownWindow {
			return nil
		}
		why = fmt.Sprintf("this session last showed it %s ago, more than %s", age.Round(time.Minute), ShownWindow)
	}
	return NewError(CodeUnread,
		"refusing to write to %s: %s. Run 'ant show %s' first and read it; if you genuinely know what you are overwriting, pass %s.",
		publicID, why, publicID, SkipReadCheckFlag)
}

// checkBodyShrink refuses a replacement body that is less than half the
// length of the existing one, unless force is set.
func checkBodyShrink(publicID, existing, replacement string, force bool) error {
	oldLen := len(strings.TrimSpace(existing))
	newLen := len(strings.TrimSpace(replacement))
	if force || oldLen < ShrinkGuardMinLength || newLen*2 >= oldLen {
		return nil
	}
	return NewError(CodeShrink,
		"refusing to replace the body of %s: the new text is %d characters, the existing one is %d. That looks like a diff or a summary rather than the whole body. If this is a deliberate rewrite, pass --force.",
		publicID, newLen, oldLen)
}
