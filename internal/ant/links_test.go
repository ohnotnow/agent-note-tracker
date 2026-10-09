package ant

import (
	"reflect"
	"strings"
	"testing"
)

// supersedePair adds an older entry and a newer one that supersedes it.
func supersedePair(t *testing.T, ta *testApp, reason string) (older, newer string) {
	t.Helper()
	older = addAndID(t, ta, "--title", "First-wave design")
	args := []string{"--title", "Second-wave design", "--supersedes", older}
	if reason != "" {
		args = append(args, "--supersedes-reason", reason)
	}
	newer = addAndID(t, ta, args...)
	return older, newer
}

func TestSupersedes_ShowBothEnds(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	older, newer := supersedePair(t, ta, "one story per folder")

	t.Run("older entry leads with superseded_by", func(t *testing.T) {
		var got EntryShown
		ta.run(t, &got, "show", older)
		want := []EntryRef{{PublicID: newer, Title: "Second-wave design", Reason: "one story per folder"}}
		if !reflect.DeepEqual(got.SupersededBy, want) {
			t.Errorf("superseded_by = %+v, want %+v", got.SupersededBy, want)
		}
		if !strings.HasPrefix(ta.stdoutString(), "{\n  \"superseded_by\"") {
			t.Errorf("superseded_by should be the first key:\n%s", ta.stdoutString())
		}
	})

	t.Run("newer entry lists supersedes", func(t *testing.T) {
		var got EntryShown
		ta.run(t, &got, "show", newer)
		want := []EntryRef{{PublicID: older, Title: "First-wave design", Reason: "one story per folder"}}
		if !reflect.DeepEqual(got.Supersedes, want) {
			t.Errorf("supersedes = %+v, want %+v", got.Supersedes, want)
		}
		if len(got.SupersededBy) != 0 {
			t.Errorf("newer entry should not be superseded: %+v", got.SupersededBy)
		}
	})
}

func TestSupersedes_MarkerOnListingRows(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	older, newer := supersedePair(t, ta, "")

	for _, cmd := range []string{"list", "recent"} {
		t.Run(cmd, func(t *testing.T) {
			rows := runList[EntrySlim](t, ta, cmd)
			got := map[string][]string{}
			for _, r := range rows {
				got[r.PublicID] = r.SupersededBy
			}
			if !reflect.DeepEqual(got[older], []string{newer}) {
				t.Errorf("%s: older superseded_by = %q, want [%s]", cmd, got[older], newer)
			}
			if got[newer] != nil {
				t.Errorf("%s: newer superseded_by = %q, want none", cmd, got[newer])
			}
		})
	}

	t.Run("human table", func(t *testing.T) {
		ta.reset()
		if err := ta.Dispatch("list", []string{"--human"}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(ta.stdoutString(), "First-wave design (superseded by "+newer+")") {
			t.Errorf("human table missing superseded marker:\n%s", ta.stdoutString())
		}
	})

	t.Run("markdown export", func(t *testing.T) {
		ta.reset()
		if err := ta.Dispatch("export", []string{older}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(ta.stdoutString(), "superseded_by: "+newer+"\n") {
			t.Errorf("export missing superseded_by:\n%s", ta.stdoutString())
		}
	})
}

func TestSupersedes_AddValidation(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	t.Run("unknown target refused and nothing added", func(t *testing.T) {
		err := ta.Dispatch("add", []string{"--body", "x", "--supersedes", "demo-ZZZZZ"})
		if err == nil || !strings.Contains(err.Error(), "demo-ZZZZZ") {
			t.Fatalf("expected an error naming the missing id, got %v", err)
		}
		if rows := runList[EntrySlim](t, ta, "list"); len(rows) != 0 {
			t.Errorf("entry was added despite the bad link: %+v", rows)
		}
	})

	t.Run("reason without supersedes refused", func(t *testing.T) {
		err := ta.Dispatch("add", []string{"--body", "x", "--supersedes-reason", "why"})
		if err == nil {
			t.Fatal("expected a usage error")
		}
	})
}

func TestSupersedes_Edit(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	older := addAndID(t, ta, "--title", "Old")
	newer := addAndID(t, ta, "--title", "New")

	supersededBy := func() []EntryRef {
		t.Helper()
		var got EntryShown
		ta.run(t, &got, "show", older)
		return got.SupersededBy
	}

	ta.run(t, nil, "edit", newer, "--supersedes", older)
	if got := supersededBy(); len(got) != 1 || got[0].PublicID != newer || got[0].Reason != "" {
		t.Fatalf("after edit --supersedes: %+v", got)
	}

	ta.run(t, nil, "edit", newer, "--supersedes", older, "--supersedes-reason", "only the parsing part")
	if got := supersededBy(); len(got) != 1 || got[0].Reason != "only the parsing part" {
		t.Fatalf("re-declaring should update the reason, got %+v", got)
	}

	ta.run(t, nil, "edit", newer, "--unsupersede", older)
	if got := supersededBy(); len(got) != 0 {
		t.Fatalf("after --unsupersede: %+v", got)
	}

	if err := ta.Dispatch("edit", []string{newer, "--unsupersede", older}); err == nil {
		t.Error("removing a link that does not exist should error")
	}
	if err := ta.Dispatch("edit", []string{newer, "--supersedes", newer}); err == nil {
		t.Error("an entry superseding itself should be refused")
	}
}

func TestSupersedes_DeleteRemovesLinks(t *testing.T) {
	t.Run("delete the newer entry", func(t *testing.T) {
		ta := newTestApp(t)
		initDemo(t, ta, "demo")
		older, newer := supersedePair(t, ta, "")
		ta.run(t, nil, "delete", "--force", newer)

		var got EntryShown
		ta.run(t, &got, "show", older)
		if len(got.SupersededBy) != 0 {
			t.Errorf("older still superseded by a deleted entry: %+v", got.SupersededBy)
		}
	})

	t.Run("delete the older entry", func(t *testing.T) {
		ta := newTestApp(t)
		initDemo(t, ta, "demo")
		older, newer := supersedePair(t, ta, "")
		ta.run(t, nil, "delete", "--force", older)

		var got EntryShown
		ta.run(t, &got, "show", newer)
		if len(got.Supersedes) != 0 {
			t.Errorf("newer still supersedes a deleted entry: %+v", got.Supersedes)
		}
	})
}
