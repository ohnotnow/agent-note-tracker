package ant

import (
	"strings"
	"testing"
)

func TestFoundation_NoEntry(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	err := ta.Dispatch("foundation", nil)
	if err == nil {
		t.Fatal("expected error when no foundation has been recorded")
	}
	if !strings.Contains(err.Error(), "no foundation recorded") {
		t.Errorf("error = %v, want 'no foundation recorded' substring", err)
	}
}

func TestFoundation_ReturnsFullEntry(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add",
		"--kind", KindFoundation,
		"--title", "Vision",
		"--body", "ant captures the why",
	)

	var got Entry
	ta.run(t, &got, "foundation")
	if got.Kind != KindFoundation {
		t.Errorf("kind = %q, want %q", got.Kind, KindFoundation)
	}
	if got.Title != "Vision" {
		t.Errorf("title = %q, want 'Vision'", got.Title)
	}
	if got.Body != "ant captures the why" {
		t.Errorf("body = %q, want 'ant captures the why'", got.Body)
	}
}

func TestFoundation_RejectsArgs(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("foundation", []string{"unexpected"}); err == nil {
		t.Error("expected error for positional argument")
	}
}

func TestAdd_FoundationSingleton(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var first Entry
	ta.run(t, &first, "add", "--kind", KindFoundation, "--body", "first")

	err := ta.Dispatch("add", []string{"--kind", KindFoundation, "--body", "second"})
	if err == nil {
		t.Fatal("expected error when adding a second foundation entry")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %v, want 'already exists' substring", err)
	}
	if !strings.Contains(err.Error(), first.PublicID) {
		t.Errorf("error = %v, want it to mention the existing id %q", err, first.PublicID)
	}
}

func TestAdd_NonFoundationKindsUnaffected(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	ta.run(t, nil, "add", "--kind", KindFoundation, "--body", "vision")
	ta.run(t, nil, "add", "--kind", KindNote, "--body", "note one")
	ta.run(t, nil, "add", "--kind", KindNote, "--body", "note two")
	ta.run(t, nil, "add", "--kind", KindADR, "--body", "adr one")
}
