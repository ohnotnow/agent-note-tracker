package ant

import (
	"strings"
	"testing"
)

func TestHandover_NoEntry(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--body", "a note, not a handover")

	err := ta.Dispatch("handover", nil)
	if err == nil {
		t.Fatal("expected error when no handover has been recorded")
	}
	if !strings.Contains(err.Error(), "no handover recorded") {
		t.Errorf("error = %v, want 'no handover recorded' substring", err)
	}
}

func TestHandoverCreate_FirstHasNoSupersedes(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got Entry
	ta.run(t, &got, "handover", "create", "--title", "Day 1", "--body", "where we are")
	if got.Kind != KindHandover {
		t.Errorf("kind = %q, want %q", got.Kind, KindHandover)
	}

	var shown EntryShown
	ta.run(t, &shown, "show", got.PublicID)
	if len(shown.Supersedes) != 0 {
		t.Errorf("supersedes = %v, want none for the first handover", shown.Supersedes)
	}
}

func TestHandoverCreate_SupersedesPrevious(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var first, second Entry
	ta.run(t, &first, "handover", "create", "--title", "Day 1", "--body", "day one")
	ta.run(t, nil, "add", "--kind", KindNote, "--body", "an unrelated note in between")
	ta.run(t, &second, "handover", "create", "--title", "Day 2", "--body", "day two")

	var shown EntryShown
	ta.run(t, &shown, "handover")
	if shown.PublicID != second.PublicID {
		t.Errorf("handover = %q, want the newest %q", shown.PublicID, second.PublicID)
	}
	if shown.Body != "day two" {
		t.Errorf("body = %q, want 'day two'", shown.Body)
	}
	if len(shown.Supersedes) != 1 || shown.Supersedes[0].PublicID != first.PublicID {
		t.Errorf("supersedes = %v, want just %q", shown.Supersedes, first.PublicID)
	}

	var old EntryShown
	ta.run(t, &old, "show", first.PublicID)
	if len(old.SupersededBy) != 1 || old.SupersededBy[0].PublicID != second.PublicID {
		t.Errorf("superseded_by = %v, want just %q", old.SupersededBy, second.PublicID)
	}
}

func TestHandover_NewestWinsWithoutLinks(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	ta.run(t, nil, "handover", "create", "--body", "linked")
	var unlinked Entry
	ta.run(t, &unlinked, "add", "--kind", KindHandover, "--body", "added by hand, no link")

	var shown EntryShown
	ta.run(t, &shown, "handover")
	if shown.PublicID != unlinked.PublicID {
		t.Errorf("handover = %q, want the newest %q", shown.PublicID, unlinked.PublicID)
	}
}

func TestHandoverCreate_RefusesKindAndSupersedes(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	var note Entry
	ta.run(t, &note, "add", "--body", "a note")

	for _, args := range [][]string{
		{"--kind", "note", "--body", "x"},
		{"--type=note", "--body", "x"},
		{"--supersedes", note.PublicID, "--body", "x"},
	} {
		err := ta.Dispatch("handover", append([]string{"create"}, args...))
		if err == nil {
			t.Errorf("handover create %v: expected a usage error", args)
		}
	}

	var list listResp[EntrySlim]
	ta.run(t, &list, "list", "--kind", KindHandover)
	if len(list.Entries) != 0 {
		t.Errorf("refused creates left %d handover entries behind", len(list.Entries))
	}
}

func TestHandover_RejectsArgs(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("handover", []string{"unexpected"}); err == nil {
		t.Error("expected error for an unknown positional argument")
	}
}
