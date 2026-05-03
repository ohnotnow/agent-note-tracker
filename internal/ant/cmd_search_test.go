package ant

import (
	"strings"
	"testing"
)

func seedSearchable(t *testing.T, ta *testApp) {
	t.Helper()
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--title", "Choose sqlite", "--body", "we picked modernc/sqlite over CGO bindings")
	ta.run(t, nil, "add", "--body", "auth flow refactored to use signed cookies", "--kind", "pivot")
	ta.run(t, nil, "add", "--body", "small bug fix unrelated to anything")
}

func TestSearch_SingleTerm(t *testing.T) {
	ta := newTestApp(t)
	seedSearchable(t, ta)

	got := runList[EntryWithSnippet](t, ta, "search", "sqlite")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if !strings.Contains(strings.ToLower(got[0].Snippet+got[0].Title), "sqlite") {
		t.Errorf("expected sqlite in title or snippet: %+v", got[0])
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	ta := newTestApp(t)
	seedSearchable(t, ta)

	got := runList[EntryWithSnippet](t, ta, "search", "SQLITE")
	if len(got) != 1 {
		t.Errorf("expected 1 result for case-insensitive 'SQLITE', got %d", len(got))
	}
}

func TestSearch_MatchesTitle(t *testing.T) {
	ta := newTestApp(t)
	seedSearchable(t, ta)

	// "Choose" only appears in a title.
	got := runList[EntryWithSnippet](t, ta, "search", "choose")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestSearch_MultiTerm_AND(t *testing.T) {
	ta := newTestApp(t)
	seedSearchable(t, ta)

	got := runList[EntryWithSnippet](t, ta, "search", "auth", "cookies")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (only the entry with both terms)", len(got))
	}

	got = runList[EntryWithSnippet](t, ta, "search", "auth", "sqlite") // no entry has both
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestSearch_QuotedPhraseSplitsOnWhitespace(t *testing.T) {
	ta := newTestApp(t)
	seedSearchable(t, ta)

	// Single arg with whitespace splits to two terms internally
	got := runList[EntryWithSnippet](t, ta, "search", "auth flow")
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestSearch_RequiresQuery(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("search", nil); err == nil {
		t.Error("expected error for empty query")
	}
}
