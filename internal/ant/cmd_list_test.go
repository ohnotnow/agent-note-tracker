package ant

import (
	"strings"
	"testing"
)

func seedEntries(t *testing.T, ta *testApp) {
	t.Helper()
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--body", "alpha", "--kind", "note")
	ta.run(t, nil, "add", "--body", "bravo", "--kind", "adr", "--issue", "ait-Foo")
	ta.run(t, nil, "add", "--body", "charlie", "--kind", "pivot", "--issue", "ait-Foo")
}

func TestList_DefaultSlim(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	got := runList[EntrySlim](t, ta, "list")
	if len(got) != 3 {
		t.Fatalf("len(list) = %d, want 3", len(got))
	}
	// Ordered DESC by created_at, but we created in <1s so id DESC tiebreaks.
	// The most recent id should be first.
	if !strings.HasPrefix(got[0].PublicID, "demo-") {
		t.Errorf("public_id missing prefix: %q", got[0].PublicID)
	}
}

func TestList_LongIncludesBody(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	got := runList[Entry](t, ta, "list", "--long")
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for _, e := range got {
		if e.Body == "" {
			t.Errorf("entry %s has empty body in --long output", e.PublicID)
		}
	}
}

func TestList_FilterKind(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	got := runList[EntrySlim](t, ta, "list", "--kind", "adr")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Kind != "adr" {
		t.Errorf("kind = %q, want adr", got[0].Kind)
	}
}

func TestList_FilterIssue(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	got := runList[EntrySlim](t, ta, "list", "--issue", "ait-Foo")
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for _, e := range got {
		if e.IssueID != "ait-Foo" {
			t.Errorf("issue_id = %q, want ait-Foo", e.IssueID)
		}
	}
}

func TestList_HumanIsTabular(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	if err := ta.Dispatch("list", []string{"--human"}); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	out := ta.stdoutString()
	if !strings.Contains(out, "DATE") || !strings.Contains(out, "KIND") || !strings.Contains(out, "ID") || !strings.Contains(out, "TITLE") {
		t.Errorf("expected header columns in human output:\n%s", out)
	}
}

func TestList_LongAndHuman_AreExclusive(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("list", []string{"--long", "--human"}); err == nil {
		t.Error("expected error for combining --long and --human")
	}
}

func TestList_SinceFiltersEarly(t *testing.T) {
	ta := newTestApp(t)
	seedEntries(t, ta)

	// 'now' lives inside the seed window; pick a future date
	got := runList[EntrySlim](t, ta, "list", "--since", "2099-01-01")
	if len(got) != 0 {
		t.Errorf("expected empty result for far-future --since, got %d", len(got))
	}
}

func TestList_RejectsBadSince(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("list", []string{"--since", "notadate"}); err == nil {
		t.Error("expected error for unparseable --since")
	}
}

func TestList_EmptyDBReturnsEmptyArray(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	got := runList[EntrySlim](t, ta, "list")
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d entries", len(got))
	}
	if !strings.Contains(ta.stdoutString(), `"entries": []`) {
		t.Errorf("expected envelope with empty array, got:\n%s", ta.stdoutString())
	}
}
