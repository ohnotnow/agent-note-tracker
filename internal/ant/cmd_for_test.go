package ant

import "testing"

func TestFor_ExactMatchOnly(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--body", "x", "--issue", "ait-Foo")
	ta.run(t, nil, "add", "--body", "y", "--issue", "ait-Foo.1")
	ta.run(t, nil, "add", "--body", "z", "--issue", "")

	var got []EntrySlim
	ta.run(t, &got, "for", "ait-Foo")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (no partial-prefix matching)", len(got))
	}
	if got[0].IssueID != "ait-Foo" {
		t.Errorf("issue_id = %q, want ait-Foo", got[0].IssueID)
	}
}

func TestFor_NoMatchReturnsEmpty(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--body", "x", "--issue", "ait-Foo")

	var got []EntrySlim
	ta.run(t, &got, "for", "ait-Bar")
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestFor_RequiresOneArg(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("for", nil); err == nil {
		t.Error("expected usage error with no arg")
	}
}
