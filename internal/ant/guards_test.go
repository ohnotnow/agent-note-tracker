package ant

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain clears CLAUDE_CODE_SESSION_ID so the read check (which only bites
// when a session id is present) stays out of tests that are not about it.
// The read-check tests set the variable explicitly with t.Setenv.
func TestMain(m *testing.M) {
	os.Unsetenv("CLAUDE_CODE_SESSION_ID")
	os.Exit(m.Run())
}

func expectCode(t *testing.T, err error, code string) *CodedError {
	t.Helper()
	var ce *CodedError
	if !errors.As(err, &ce) || ce.Code != code {
		t.Fatalf("expected %s error, got %v", code, err)
	}
	return ce
}

func TestEdit_RefusesEmptyBody(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "precious body")

	// Explicit stdin with nothing on it: the empty-heredoc mistake.
	ta.Stdin = strings.NewReader("")
	expectCode(t, ta.Dispatch("edit", []string{id, "--body", "-"}), CodeValidationError)
	expectCode(t, ta.Dispatch("edit", []string{id, "--body", "   \n"}), CodeValidationError)

	var got Entry
	ta.run(t, &got, "show", id)
	if got.Body != "precious body" {
		t.Fatalf("body was changed: %q", got.Body)
	}
}

func TestEdit_RefusesLargeShrinkWithoutForce(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	long := strings.Repeat("A long and detailed body. ", 12)
	id := addAndID(t, ta, "--body", long)

	ce := expectCode(t, ta.Dispatch("edit", []string{id, "--body", "short summary"}), CodeShrink)
	if !strings.Contains(ce.Message, "--force") {
		t.Fatalf("refusal should name --force: %s", ce.Message)
	}

	// Growing, or a modest trim, is never questioned.
	ta.run(t, nil, "edit", id, "--body", long+"One more sentence.")
	ta.run(t, nil, "edit", id, "--body", strings.Repeat("A long and detailed body. ", 8))

	// A short existing body can be replaced by anything.
	small := addAndID(t, ta, "--body", "todo: write this up")
	ta.run(t, nil, "edit", small, "--body", "x")

	// --force allows the deliberate rewrite.
	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--body", "short summary", "--force")
	if got.Body != "short summary" {
		t.Fatalf("forced rewrite not applied: %q", got.Body)
	}
}

func TestReadCheck_IgnoredWithoutSessionID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "human edit")
	ta.run(t, nil, "edit", id, "--title", "renamed without a show")
	ta.run(t, nil, "edit", id, "--body", "rewritten without a show")
}

func TestReadCheck_RefusesUnshownEntryInSession(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-a")
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "agent edit")

	for _, args := range [][]string{
		{id, "--title", "blind rename"},
		{id, "--body", "blind rewrite"},
	} {
		ce := expectCode(t, ta.Dispatch("edit", args), CodeUnread)
		if !strings.Contains(ce.Message, SkipReadCheckFlag) {
			t.Fatalf("refusal should name the override flag: %s", ce.Message)
		}
	}

	// Kind and issue are metadata, not body writes, and pass.
	ta.run(t, nil, "edit", id, "--kind", "adr", "--issue", "ait-AbCdE")
	// So does append, which only ever grows the body.
	ta.run(t, nil, "append", id, "--body", "more")

	// The override flag works.
	ta.run(t, nil, "edit", id, "--title", "forced rename", SkipReadCheckFlag)
}

func TestReadCheck_PassesAfterShowInSameSession(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-a")
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "agent edit")
	ta.run(t, nil, "show", id)
	ta.run(t, nil, "edit", id, "--title", "read then renamed")
	ta.run(t, nil, "edit", id, "--body", "read then rewritten")
}

func TestReadCheck_RefusesShowFromAnotherSession(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-a")
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "shared")
	ta.run(t, nil, "show", id)

	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-b")
	ce := expectCode(t, ta.Dispatch("edit", []string{id, "--title", "from b"}), CodeUnread)
	if !strings.Contains(ce.Message, "by a different session") {
		t.Fatalf("expected the different-session reason, got: %s", ce.Message)
	}
}

func TestReadCheck_ExpiresAfterWindow(t *testing.T) {
	t.Setenv("CLAUDE_CODE_SESSION_ID", "session-a")
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "stale")
	ta.run(t, nil, "show", id)

	store, _, err := ta.requireInitialised()
	if err != nil {
		t.Fatal(err)
	}
	stale := time.Now().UTC().Add(-ShownWindow - time.Minute).Format(time.RFC3339)
	if _, err := store.DB.Exec(`UPDATE entries SET shown_at = ? WHERE public_id = ?`, stale, id); err != nil {
		t.Fatal(err)
	}

	ce := expectCode(t, ta.Dispatch("edit", []string{id, "--title", "too late"}), CodeUnread)
	if !strings.Contains(ce.Message, "more than") {
		t.Fatalf("expected the expiry reason, got: %s", ce.Message)
	}

	ta.run(t, nil, "show", id)
	ta.run(t, nil, "edit", id, "--title", "fresh look")
}
