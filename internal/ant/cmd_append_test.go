package ant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppend_FlagText(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	ta.run(t, nil, "append", id, "--body", "second beat")

	var got Entry
	ta.run(t, &got, "show", id)
	want := "original body\n\n---\n\nsecond beat"
	if got.Body != want {
		t.Errorf("body =\n%q\nwant\n%q", got.Body, want)
	}
}

func TestAppend_DefaultIsSlim(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	var got map[string]any
	ta.run(t, &got, "append", id, "--body", "more")

	if _, hasBody := got["body"]; hasBody {
		t.Errorf("default append response includes body — should be slim:\n%s", ta.stdoutString())
	}
}

func TestAppend_LongReturnsFullEntry(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	var got Entry
	ta.run(t, &got, "append", id, "--long", "--body", "more")
	if !strings.Contains(got.Body, "original body") || !strings.Contains(got.Body, "more") {
		t.Errorf("body = %q, expected both original and appended content", got.Body)
	}
}

func TestAppend_FromStdin(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.Stdin = strings.NewReader("piped addition\n")

	ta.run(t, nil, "append", id)

	var got Entry
	ta.run(t, &got, "show", id)
	if !strings.HasSuffix(got.Body, "---\n\npiped addition") {
		t.Errorf("body = %q, expected suffix '---\\n\\npiped addition'", got.Body)
	}
}

func TestAppend_ExplicitStdinFlag(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.Stdin = strings.NewReader("explicit stdin\n")

	ta.run(t, nil, "append", id, "--body", "-")

	var got Entry
	ta.run(t, &got, "show", id)
	if !strings.HasSuffix(got.Body, "---\n\nexplicit stdin") {
		t.Errorf("body = %q, expected suffix '---\\n\\nexplicit stdin'", got.Body)
	}
}

func TestAppend_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "more.md")
	if err := os.WriteFile(path, []byte("from a file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	ta.run(t, nil, "append", id, "--body", "@"+path)

	var got Entry
	ta.run(t, &got, "show", id)
	if !strings.HasSuffix(got.Body, "---\n\nfrom a file") {
		t.Errorf("body = %q, expected the file content appended after a divider", got.Body)
	}
}

func TestAppend_MultipleAppendsChain(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	ta.run(t, nil, "append", id, "--body", "second")
	ta.run(t, nil, "append", id, "--body", "third")

	var got Entry
	ta.run(t, &got, "show", id)
	want := "original body\n\n---\n\nsecond\n\n---\n\nthird"
	if got.Body != want {
		t.Errorf("body =\n%q\nwant\n%q", got.Body, want)
	}
}

func TestAppend_RejectsEmptyContent(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	if err := ta.Dispatch("append", []string{id}); err == nil {
		t.Error("expected error when there is nothing to append")
	}
}

func TestAppend_RejectsWhitespaceOnlyContent(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	if err := ta.Dispatch("append", []string{id, "--body", "   \n\n"}); err == nil {
		t.Error("expected error when appended content is whitespace-only")
	}
}

func TestAppend_MissingID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	if err := ta.Dispatch("append", []string{"demo-doesnotexist", "--body", "x"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestAppend_RequiresOneArg(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("append", []string{"--body", "x"}); err == nil {
		t.Error("expected usage error with no positional arg")
	}
}

func TestAppend_RequiresInit(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("append", []string{"demo-AbCdE", "--body", "x"})
	if err == nil {
		t.Fatal("expected error before init")
	}
	if !strings.Contains(err.Error(), "ant init") {
		t.Errorf("error = %v, want hint to run 'ant init'", err)
	}
}

func TestAppend_PreservesTitleAndKind(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	var first Entry
	ta.run(t, &first, "add", "--long", "--body", "rationale",
		"--title", "Use sqlite", "--kind", "adr", "--issue", "ait-AbCdE.1")

	ta.run(t, nil, "append", first.PublicID, "--body", "follow-up")

	var got Entry
	ta.run(t, &got, "show", first.PublicID)
	if got.Title != "Use sqlite" || got.Kind != "adr" || got.IssueID != "ait-AbCdE.1" {
		t.Errorf("metadata changed after append: %+v", got)
	}
}
