package ant

import (
	"strings"
	"testing"
)

func TestRecent_DefaultLimit5(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	for i := 0; i < 8; i++ {
		ta.run(t, nil, "add", "--body", "body")
	}

	var got []EntryWithSnippet
	ta.run(t, &got, "recent")
	if len(got) != 5 {
		t.Fatalf("len = %d, want 5", len(got))
	}
}

func TestRecent_CustomLimit(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	for i := 0; i < 4; i++ {
		ta.run(t, nil, "add", "--body", "body")
	}

	var got []EntryWithSnippet
	ta.run(t, &got, "recent", "--limit", "2")
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestRecent_IncludesSnippet(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	long := strings.Repeat("x", 500)
	ta.run(t, nil, "add", "--body", long)

	var got []EntryWithSnippet
	ta.run(t, &got, "recent")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if !strings.HasSuffix(got[0].Snippet, "…") {
		t.Errorf("snippet should end with ellipsis when truncated: %q", got[0].Snippet)
	}
	// Snippet has SnippetLen runes plus the ellipsis.
	if len([]rune(got[0].Snippet)) != SnippetLen+1 {
		t.Errorf("snippet rune length = %d, want %d", len([]rune(got[0].Snippet)), SnippetLen+1)
	}
}

func TestRecent_ShortBodyHasNoEllipsis(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--body", "short")

	var got []EntryWithSnippet
	ta.run(t, &got, "recent")
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Snippet != "short" {
		t.Errorf("snippet = %q, want 'short'", got[0].Snippet)
	}
}

func TestRecent_RejectsNonPositiveLimit(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	if err := ta.Dispatch("recent", []string{"--limit", "0"}); err == nil {
		t.Error("expected error for --limit 0")
	}
}
