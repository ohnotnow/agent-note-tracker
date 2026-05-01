package ant

import (
	"strings"
	"testing"
)

func TestExport_SingleEntry_Markdown(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--title", "Picked sqlite", "--kind", "adr", "--issue", "ait-Foo")

	ta.reset()
	if err := ta.Dispatch("export", []string{id}); err != nil {
		t.Fatalf("export: %v", err)
	}
	out := ta.stdoutString()

	for _, want := range []string{
		"---\n",
		"id: " + id,
		"kind: adr",
		"issue_id: ait-Foo",
		"created_at:",
		"## Picked sqlite",
		"original body",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("export output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "***") {
		t.Errorf("single-entry export should have no separator:\n%s", out)
	}
}

func TestExport_FilteredSet_HasSeparator(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	addAndID(t, ta, "--kind", "adr")
	addAndID(t, ta, "--kind", "adr")
	addAndID(t, ta, "--kind", "note")

	ta.reset()
	if err := ta.Dispatch("export", []string{"--kind", "adr"}); err != nil {
		t.Fatalf("export: %v", err)
	}
	out := ta.stdoutString()

	if !strings.Contains(out, "***") {
		t.Errorf("multi-entry export missing horizontal-rule separator:\n%s", out)
	}
	// Both entries' frontmatter should be present.
	if strings.Count(out, "kind: adr") != 2 {
		t.Errorf("expected 2 'kind: adr' frontmatter lines, got %d:\n%s",
			strings.Count(out, "kind: adr"), out)
	}
	// The note kind entry should not be present.
	if strings.Contains(out, "kind: note") {
		t.Errorf("expected adr-only export but found 'kind: note':\n%s", out)
	}
}

func TestExport_TitleFallsBackToID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta) // no --title

	ta.reset()
	if err := ta.Dispatch("export", []string{id}); err != nil {
		t.Fatalf("export: %v", err)
	}
	out := ta.stdoutString()
	if !strings.Contains(out, "## "+id) {
		t.Errorf("expected H2 heading with id when no title:\n%s", out)
	}
}

func TestExport_JSON(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	addAndID(t, ta, "--title", "First")
	addAndID(t, ta, "--title", "Second")

	var got []Entry
	ta.run(t, &got, "export", "--json")
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestExport_RejectsIDWithFilters(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	if err := ta.Dispatch("export", []string{id, "--kind", "adr"}); err == nil {
		t.Error("expected error when combining id with filter")
	}
}

func TestExport_MissingID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("export", []string{"demo-doesnotexist"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestExport_EmptyDB_Markdown(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.reset()
	if err := ta.Dispatch("export", nil); err != nil {
		t.Fatalf("export: %v", err)
	}
	if ta.stdoutString() != "" {
		t.Errorf("expected empty output for empty DB, got %q", ta.stdoutString())
	}
}
