package ant

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func addAndID(t *testing.T, ta *testApp, args ...string) string {
	t.Helper()
	all := append([]string{"--body", "original body"}, args...)
	var e Entry
	ta.run(t, &e, "add", all...)
	return e.PublicID
}

func TestEdit_BodyViaFlag(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--body", "new body text")
	if got.Body != "new body text" {
		t.Errorf("body = %q, want 'new body text'", got.Body)
	}
}

func TestEdit_BodyViaStdin(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.Stdin = strings.NewReader("piped new body\n")

	var got Entry
	ta.run(t, &got, "edit", id, "--long")
	if got.Body != "piped new body" {
		t.Errorf("body = %q, want 'piped new body'", got.Body)
	}
}

func TestEdit_BodyExplicitStdinFlag(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.Stdin = strings.NewReader("via --body -\n")

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--body", "-")
	if got.Body != "via --body -" {
		t.Errorf("body = %q, want 'via --body -'", got.Body)
	}
}

func TestEdit_BodyFlagBeatsStdin(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.Stdin = strings.NewReader("from stdin")

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--body", "from flag")
	if got.Body != "from flag" {
		t.Errorf("body = %q, want 'from flag'", got.Body)
	}
}

func TestEdit_TitleSetAndCleared(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--title", "Original")

	var setRes Entry
	ta.run(t, &setRes, "edit", id, "--title", "Updated")
	if setRes.Title != "Updated" {
		t.Errorf("title = %q, want 'Updated'", setRes.Title)
	}

	var clearRes Entry
	ta.run(t, &clearRes, "edit", id, "--title", "")
	if clearRes.Title != "" {
		t.Errorf("title = %q, want '' after clear", clearRes.Title)
	}
}

func TestEdit_IssueSetAndCleared(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--issue", "ait-Foo")

	var setRes Entry
	ta.run(t, &setRes, "edit", id, "--issue", "ait-Bar")
	if setRes.IssueID != "ait-Bar" {
		t.Errorf("issue_id = %q, want ait-Bar", setRes.IssueID)
	}

	var clearRes Entry
	ta.run(t, &clearRes, "edit", id, "--issue", "")
	if clearRes.IssueID != "" {
		t.Errorf("issue_id = %q, want '' after clear", clearRes.IssueID)
	}
}

func TestEdit_KindUpdate(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--kind", "note")

	var got Entry
	ta.run(t, &got, "edit", id, "--kind", "adr")
	if got.Kind != "adr" {
		t.Errorf("kind = %q, want adr", got.Kind)
	}
}

func TestEdit_RejectsEmptyKind(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--kind", "note")

	if err := ta.Dispatch("edit", []string{id, "--kind", ""}); err == nil {
		t.Error("expected error when --kind is empty")
	}
}

func TestEdit_NoFlagsIsNoop(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--title", "T", "--kind", "note")

	var got Entry
	ta.run(t, &got, "edit", id)
	if got.Title != "T" || got.Kind != "note" {
		t.Errorf("noop edit changed something: %+v", got)
	}
}

func TestEdit_MissingID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("edit", []string{"demo-doesnotexist", "--body", "x"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestEdit_RequiresOneArg(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("edit", []string{"--body", "x"}); err == nil {
		t.Error("expected usage error with no positional arg")
	}
}

// fakeEditorScript writes a small shell script that overwrites its argument
// with the given content and returns its path. Used to drive --visual tests
// without needing a real interactive editor.
func fakeEditorScript(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-editor.sh")
	body := "#!/bin/sh\nprintf %s " + shellQuote(content) + " > \"$1\"\n"
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake editor: %v", err)
	}
	return path
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func TestEdit_VisualEditor_UpdatesBody(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("--visual test uses a POSIX shell script as the fake editor")
	}
	t.Setenv("EDITOR", fakeEditorScript(t, "edited via fake editor"))

	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "original body")

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--visual")
	if got.Body != "edited via fake editor" {
		t.Errorf("body = %q, want 'edited via fake editor'", got.Body)
	}
}

func TestEdit_VisualEditor_RejectsEmptyResult(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("--visual test uses a POSIX shell script as the fake editor")
	}
	t.Setenv("EDITOR", fakeEditorScript(t, ""))

	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "original")

	if err := ta.Dispatch("edit", []string{id, "--visual"}); err == nil {
		t.Error("expected error when editor produces an empty body")
	}
}

func TestEdit_VisualEditor_NoChangeIsNoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("--visual test uses a POSIX shell script as the fake editor")
	}
	t.Setenv("EDITOR", fakeEditorScript(t, "untouched"))

	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "untouched")

	var before Entry
	ta.run(t, &before, "show", id)

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--visual")
	if got.UpdatedAt != before.UpdatedAt {
		t.Errorf("updated_at changed despite no body change: before=%q after=%q",
			before.UpdatedAt, got.UpdatedAt)
	}
}

func TestEdit_VisualAndBody_Conflict(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--body", "x")
	if err := ta.Dispatch("edit", []string{id, "--visual", "--body", "y"}); err == nil {
		t.Error("expected error when --visual and --body are combined")
	}
}
