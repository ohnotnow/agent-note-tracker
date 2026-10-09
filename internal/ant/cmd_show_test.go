package ant

import (
	"reflect"
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

func TestShow_Mentions(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	older := addAndID(t, ta, "--title", "First-wave design")
	newer := addAndID(t, ta, "--title", "Second-wave design",
		"--body", "Builds on "+older+"; see also demo-ZZZZZ and ait-AbCdE.")

	t.Run("mentions and ids not in the notebook", func(t *testing.T) {
		var got EntryShown
		ta.run(t, &got, "show", newer)
		want := []EntryRef{{PublicID: older, Title: "First-wave design"}}
		if !reflect.DeepEqual(got.Mentions, want) {
			t.Errorf("mentions = %+v, want %+v", got.Mentions, want)
		}
		if !reflect.DeepEqual(got.NotInNotebook, []string{"demo-ZZZZZ"}) {
			t.Errorf("not_in_notebook = %q, want [demo-ZZZZZ]", got.NotInNotebook)
		}
		if len(got.MentionedBy) != 0 {
			t.Errorf("mentioned_by = %+v, want none", got.MentionedBy)
		}
	})

	t.Run("mentioned_by on the older entry", func(t *testing.T) {
		var got EntryShown
		ta.run(t, &got, "show", older)
		want := []EntryRef{{PublicID: newer, Title: "Second-wave design"}}
		if !reflect.DeepEqual(got.MentionedBy, want) {
			t.Errorf("mentioned_by = %+v, want %+v", got.MentionedBy, want)
		}
	})
}

func TestShow_SelfMention_Ignored(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.run(t, nil, "edit", "--body", "This entry is "+id+".", id)

	var got EntryShown
	ta.run(t, &got, "show", id)
	if len(got.Mentions) != 0 || len(got.MentionedBy) != 0 || len(got.NotInNotebook) != 0 {
		t.Errorf("self-mention leaked into links: %+v", got)
	}
}

func TestShow_NoMentions_OmitsKeys(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	ta.run(t, nil, "show", id)
	for _, key := range []string{`"mentions"`, `"mentioned_by"`, `"not_in_notebook"`} {
		if strings.Contains(ta.stdoutString(), key) {
			t.Errorf("show output has %s with nothing to report:\n%s", key, ta.stdoutString())
		}
	}
}
