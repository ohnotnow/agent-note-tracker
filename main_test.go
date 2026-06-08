package main

import (
	"testing"

	"agent-note-tracker/internal/ant"
)

// A --db flag with no value is a CLI-grammar error: it must surface as a
// usage-class failure (exit 64), matching ait, not a generic exit 1.
func TestExtractDBFlag_MissingValueIsUsage(t *testing.T) {
	_, _, err := extractDBFlag([]string{"--db"})
	if err == nil {
		t.Fatal("expected error when --db has no value")
	}
	if got := ant.ExitCodeFor(err); got != 64 {
		t.Errorf("exit = %d, want 64 (usage)", got)
	}
}

// A well-formed --db value is extracted and stripped from the args.
func TestExtractDBFlag_ExtractsValue(t *testing.T) {
	db, rest, err := extractDBFlag([]string{"--db", "/tmp/x.db", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db != "/tmp/x.db" {
		t.Errorf("db = %q, want /tmp/x.db", db)
	}
	if len(rest) != 1 || rest[0] != "list" {
		t.Errorf("rest = %v, want [list]", rest)
	}
}
