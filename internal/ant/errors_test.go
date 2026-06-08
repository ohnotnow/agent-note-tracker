package ant

import (
	"encoding/json"
	"errors"
	"flag"
	"strings"
	"testing"
)

// errEnvelope decodes the {"error":{code,message}} envelope from stderr and
// asserts stderr contains *only* that envelope — i.e. no leaked Go flag usage
// dump precedes it (the whole-buffer Unmarshal fails if anything else is there).
func errEnvelope(t *testing.T, ta *testApp) errorPayload {
	t.Helper()
	raw := ta.stderrString()
	var env errorEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("stderr is not a clean JSON error envelope (leaked output?):\n%s", raw)
	}
	return env.Error
}

func TestExitCodeFor(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"usage", NewError(CodeUsage, "x"), 64},
		{"validation", NewError(CodeValidationError, "x"), 1},
		{"not found", NewError(CodeNotFound, "x"), 1},
		{"plain", errors.New("boom"), 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ExitCodeFor(c.err); got != c.want {
				t.Errorf("ExitCodeFor(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

// An unknown flag must yield a clean usage envelope (exit 64), with none of
// Go's "flag provided but not defined" + usage block leaking to stderr. This
// is the report's B1 (and resolves the ugly `ant list --tree` error in C3).
func TestUsage_UnknownFlagIsCleanUsageError(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.reset()

	err := ta.Dispatch("list", []string{"--tree"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if got := errEnvelope(t, ta); got.Code != CodeUsage {
		t.Errorf("code = %q, want %q", got.Code, CodeUsage)
	}
	if got := ExitCodeFor(err); got != 64 {
		t.Errorf("exit = %d, want 64", got)
	}
	// The custom usage block (PrintDefaults) must not have leaked.
	if strings.Contains(ta.stderrString(), "tabular human-readable") {
		t.Errorf("Go usage block leaked to stderr:\n%s", ta.stderrString())
	}
}

func TestUsage_UnknownCommand(t *testing.T) {
	ta := newTestApp(t)
	ta.reset()

	err := ta.Dispatch("boguscmd", nil)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if got := errEnvelope(t, ta); got.Code != CodeUsage {
		t.Errorf("code = %q, want %q", got.Code, CodeUsage)
	}
	if got := ExitCodeFor(err); got != 64 {
		t.Errorf("exit = %d, want 64", got)
	}
}

func TestUsage_MissingPositional(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.reset()

	err := ta.Dispatch("show", nil)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if got := errEnvelope(t, ta); got.Code != CodeUsage {
		t.Errorf("code = %q, want %q", got.Code, CodeUsage)
	}
	if got := ExitCodeFor(err); got != 64 {
		t.Errorf("exit = %d, want 64", got)
	}
}

func TestUsage_ExtraPositional(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	id := addAndID(t, ta)
	ta.reset()

	err := ta.Dispatch("edit", []string{id, "unexpected"})
	if err == nil {
		t.Fatal("expected error for extra positional")
	}
	if got := errEnvelope(t, ta); got.Code != CodeUsage {
		t.Errorf("code = %q, want %q", got.Code, CodeUsage)
	}
}

func TestUsage_MutuallyExclusiveFlags(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.reset()

	err := ta.Dispatch("list", []string{"--long", "--human"})
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if got := errEnvelope(t, ta); got.Code != CodeUsage {
		t.Errorf("code = %q, want %q", got.Code, CodeUsage)
	}
}

// Regression guard: genuine content-validation failures stay validation_error
// at exit 1 — they are NOT usage errors (the invocation parsed fine).
func TestUsage_EmptyBodyStaysValidation(t *testing.T) {
	ta := newTestApp(t)
	initDemo(t, ta, "demo")
	ta.reset()

	err := ta.Dispatch("add", []string{"--body", ""})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if got := errEnvelope(t, ta); got.Code != CodeValidationError {
		t.Errorf("code = %q, want %q (empty body is content, not usage)", got.Code, CodeValidationError)
	}
	if got := ExitCodeFor(err); got != 1 {
		t.Errorf("exit = %d, want 1", got)
	}
}

// Regression guard: --help still prints its usage block (to stderr) and is
// signalled as flag.ErrHelp (a clean, requested help dump — not a failure).
// The parseFlags suppression must not break this.
func TestUsage_HelpStillPrints(t *testing.T) {
	ta := newTestApp(t)
	ta.reset()

	err := ta.Dispatch("list", []string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("list --help returned %v, want flag.ErrHelp", err)
	}
	if !strings.Contains(ta.stderrString(), "usage: ant list") {
		t.Errorf("--help did not print the usage block:\n%s", ta.stderrString())
	}
}
