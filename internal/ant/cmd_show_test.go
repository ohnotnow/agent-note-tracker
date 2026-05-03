package ant

import (
	"strings"
	"testing"
)

func TestShow_RoundTripsAddedEntry(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var added Entry
	ta.run(t, &added, "add", "--long",
		"--body", "load-bearing decision",
		"--title", "Choose sqlite",
		"--kind", "adr",
		"--issue", "ait-Foo",
	)

	var shown Entry
	ta.run(t, &shown, "show", added.PublicID)

	if shown.PublicID != added.PublicID ||
		shown.Title != added.Title ||
		shown.Body != added.Body ||
		shown.Kind != added.Kind ||
		shown.IssueID != added.IssueID {
		t.Errorf("show returned %+v, want %+v", shown, added)
	}
}

func TestShow_MissingID_Errors(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	err := ta.Dispatch("show", []string{"demo-doesnotexist"})
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if !strings.Contains(err.Error(), "no entry with id") {
		t.Errorf("error = %v, want 'no entry with id' substring", err)
	}
}

func TestShow_RequiresOneArg(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("show", nil); err == nil {
		t.Error("expected usage error with zero args")
	}
	ta.reset()
	if err := ta.Dispatch("show", []string{"a", "b"}); err == nil {
		t.Error("expected usage error with two args")
	}
}

func TestShow_RequiresInit(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("show", []string{"anything"})
	if err == nil {
		t.Fatal("expected error before init")
	}
}
