package ant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func initDemo(t *testing.T, ta *testApp, prefix string) {
	t.Helper()
	ta.run(t, nil, "init", "--prefix", prefix)
}

func TestAdd_BasicBodyFlag(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got Entry
	ta.run(t, &got, "add", "--long", "--body", "Hello world")

	if !strings.HasPrefix(got.PublicID, "demo-") {
		t.Errorf("public_id = %q, want demo- prefix", got.PublicID)
	}
	if got.Kind != "note" {
		t.Errorf("kind = %q, want 'note'", got.Kind)
	}
	if got.Body != "Hello world" {
		t.Errorf("body = %q, want %q", got.Body, "Hello world")
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Error("expected timestamps to be set")
	}
}

func TestAdd_DefaultIsSlim(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got map[string]any
	ta.run(t, &got, "add", "--body", "ought to be slim")

	if _, hasBody := got["body"]; hasBody {
		t.Errorf("default add response includes body — should be slim:\n%s", ta.stdoutString())
	}
	if _, hasUpdated := got["updated_at"]; hasUpdated {
		t.Errorf("default add response includes updated_at — should be slim:\n%s", ta.stdoutString())
	}
	if got["id"] == "" {
		t.Error("slim response missing id")
	}
}

func TestAdd_AllFlags(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got Entry
	ta.run(t, &got, "add",
		"--body", "rationale",
		"--title", "Use sqlite",
		"--kind", "adr",
		"--issue", "ait-AbCdE.1",
	)

	if got.Title != "Use sqlite" {
		t.Errorf("title = %q, want 'Use sqlite'", got.Title)
	}
	if got.Kind != "adr" {
		t.Errorf("kind = %q, want 'adr'", got.Kind)
	}
	if got.IssueID != "ait-AbCdE.1" {
		t.Errorf("issue_id = %q, want 'ait-AbCdE.1'", got.IssueID)
	}
}

func TestAdd_FromStdin(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.Stdin = strings.NewReader("body from stdin\n")

	var got Entry
	ta.run(t, &got, "add", "--long")
	if got.Body != "body from stdin" {
		t.Errorf("body = %q, want 'body from stdin'", got.Body)
	}
}

func TestAdd_ExplicitStdinFlag(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.Stdin = strings.NewReader("body via --body -\n")

	var got Entry
	ta.run(t, &got, "add", "--long", "--body", "-")
	if got.Body != "body via --body -" {
		t.Errorf("body = %q, want 'body via --body -'", got.Body)
	}
}

func TestAdd_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.md")
	if err := os.WriteFile(path, []byte("# heading\n\nfile content\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got Entry
	ta.run(t, &got, "add", "--long", "--body", "@"+path)
	if !strings.Contains(got.Body, "file content") {
		t.Errorf("body = %q, want it to contain 'file content'", got.Body)
	}
}

func TestAdd_RequiresBody(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	err := ta.Dispatch("add", nil)
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestAdd_RequiresInit(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("add", []string{"--body", "x"})
	if err == nil {
		t.Fatal("expected error before init")
	}
	if !strings.Contains(err.Error(), "ant init") {
		t.Errorf("error = %v, want hint to run 'ant init'", err)
	}
}

func TestAdd_UniquePublicIDs(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	seen := make(map[string]bool)
	for i := 0; i < 5; i++ {
		var e Entry
		ta.run(t, &e, "add", "--body", "x")
		if seen[e.PublicID] {
			t.Fatalf("duplicate public_id %q", e.PublicID)
		}
		seen[e.PublicID] = true
	}
}
