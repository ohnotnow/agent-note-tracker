package ant

import (
	"errors"
	"strings"
	"testing"
)

func TestDelete_WithoutForce_RefusesAndKeepsRow(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--title", "Important note")

	err := ta.Dispatch("delete", []string{id})
	if err == nil {
		t.Fatal("expected delete without --force to error")
	}
	if !strings.Contains(ta.stderrString(), "would delete") {
		t.Errorf("stderr missing 'would delete' warning:\n%s", ta.stderrString())
	}

	// Row still present.
	store, _ := ta.Store()
	if _, err := store.GetEntry(id); err != nil {
		t.Errorf("entry should still exist: %v", err)
	}
}

func TestDelete_WithForce_RemovesRowAndEchoesIt(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta, "--title", "Goodbye")

	var got Entry
	ta.run(t, &got, "delete", "--force", "--long", id)
	if got.PublicID != id {
		t.Errorf("echoed id = %q, want %q", got.PublicID, id)
	}

	store, _ := ta.Store()
	_, err := store.GetEntry(id)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDelete_MissingID(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("delete", []string{"demo-doesnotexist", "--force"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestDelete_RequiresInit(t *testing.T) {
	ta := newTestApp(t)
	if err := ta.Dispatch("delete", []string{"x", "--force"}); err == nil {
		t.Error("expected error before init")
	}
}
