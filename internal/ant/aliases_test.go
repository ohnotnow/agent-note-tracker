package ant

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

// --- unit tests for the rewrite helper -----------------------------------

func TestCanonicaliseAliases_Rewrites(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"type to kind", []string{"--type", "adr"}, []string{"--kind", "adr"}},
		{"description to body", []string{"--description", "x"}, []string{"--body", "x"}},
		{"type with equals", []string{"--type=adr"}, []string{"--kind=adr"}},
		{"single dash", []string{"-type", "adr"}, []string{"--kind", "adr"}},
		{"canonical untouched", []string{"--kind", "adr"}, []string{"--kind", "adr"}},
		{"unrelated flags untouched", []string{"--title", "t", "--long"}, []string{"--title", "t", "--long"}},
		{"stdin marker untouched", []string{"--body", "-"}, []string{"--body", "-"}},
		{"positional untouched", []string{"demo-AbCdE", "--type", "pivot"}, []string{"demo-AbCdE", "--kind", "pivot"}},
		{"after terminator untouched", []string{"--", "--type"}, []string{"--", "--type"}},
		{"empty", nil, []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := canonicaliseAliases(c.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Join(got, "\x00") != strings.Join(c.want, "\x00") {
				t.Errorf("canonicaliseAliases(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestCanonicaliseAliases_ConflictRejected(t *testing.T) {
	cases := [][]string{
		{"--kind", "a", "--type", "b"},
		{"--type", "a", "--kind", "b"},
		{"--body", "a", "--description", "b"},
		{"--description=a", "--body=b"},
	}
	for _, in := range cases {
		t.Run(strings.Join(in, " "), func(t *testing.T) {
			_, err := canonicaliseAliases(in)
			if err == nil {
				t.Fatalf("expected conflict error for %v", in)
			}
			var ce *CodedError
			if !errors.As(err, &ce) || ce.Code != CodeUsage {
				t.Errorf("error = %v, want CodeUsage", err)
			}
		})
	}
}

func TestCanonicaliseAliases_SameFlagRepeatedIsFine(t *testing.T) {
	// Repeating the *same* spelling is not a conflict — Go's flag package
	// applies last-wins, which we leave intact.
	got, err := canonicaliseAliases([]string{"--type", "a", "--type", "b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "--kind\x00a\x00--kind\x00b"; strings.Join(got, "\x00") != want {
		t.Errorf("got %v", got)
	}
}

// --- command-level: aliases behave identically to the canonical flag ------

func TestAlias_AddTypeAndDescription(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got Entry
	ta.run(t, &got, "add", "--long", "--type", "adr", "--description", "rationale text")
	if got.Kind != "adr" {
		t.Errorf("kind = %q, want 'adr'", got.Kind)
	}
	if got.Body != "rationale text" {
		t.Errorf("body = %q, want 'rationale text'", got.Body)
	}
}

// The output key stays "kind" even when the input used --type (section C2).
func TestAlias_OutputKeyStaysKind(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	var got map[string]any
	ta.run(t, &got, "add", "--long", "--type", "adr", "--body", "b")
	if _, ok := got["kind"]; !ok {
		t.Errorf("output missing 'kind' key:\n%s", ta.stdoutString())
	}
	if _, ok := got["type"]; ok {
		t.Errorf("output leaked a 'type' key — input alias must not change output:\n%s", ta.stdoutString())
	}
}

func TestAlias_EditType(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)

	var got Entry
	ta.run(t, &got, "edit", id, "--long", "--type", "pivot")
	if got.Kind != "pivot" {
		t.Errorf("kind = %q, want 'pivot'", got.Kind)
	}
}

func TestAlias_ListTypeFilter(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--kind", "adr", "--body", "an adr")
	ta.run(t, nil, "add", "--kind", "note", "--body", "a note")

	viaType := runList[EntrySlim](t, ta, "list", "--type", "adr")
	viaKind := runList[EntrySlim](t, ta, "list", "--kind", "adr")
	if len(viaType) != 1 || len(viaKind) != 1 {
		t.Fatalf("--type filter returned %d, --kind returned %d, want 1 each", len(viaType), len(viaKind))
	}
	if viaType[0].PublicID != viaKind[0].PublicID {
		t.Errorf("--type and --kind selected different entries: %q vs %q", viaType[0].PublicID, viaKind[0].PublicID)
	}
}

func TestAlias_ExportTypeFilter(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.run(t, nil, "add", "--kind", "adr", "--body", "an adr")
	ta.run(t, nil, "add", "--kind", "note", "--body", "a note")

	viaType := runList[Entry](t, ta, "export", "--type", "adr", "--json")
	if len(viaType) != 1 || viaType[0].Kind != "adr" {
		t.Fatalf("export --type adr returned %d entries, want 1 adr", len(viaType))
	}
}

func TestAlias_BothFormsRejected(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")

	ta.reset()
	err := ta.Dispatch("add", []string{"--kind", "adr", "--type", "note", "--body", "b"})
	if err == nil {
		t.Fatal("expected error when both --kind and --type are given")
	}
	var ce *CodedError
	if !errors.As(err, &ce) || ce.Code != CodeUsage {
		t.Errorf("error = %v, want CodeUsage", err)
	}
}

// Help text must not leak the input aliases — they're input-only.
func TestAlias_NotInHelpText(t *testing.T) {
	ta := newTestApp(t)
	ta.reset()
	err := ta.Dispatch("add", []string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("add --help returned %v, want flag.ErrHelp", err)
	}
	help := ta.stderrString()
	for _, leak := range []string{"type", "description"} {
		if strings.Contains(help, leak) {
			t.Errorf("help text leaked input alias %q:\n%s", leak, help)
		}
	}
}
