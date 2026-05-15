package ant

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// testApp wraps an *App with captured stdout/stderr buffers. Tests construct
// one per case via newTestApp and dispatch commands directly through it,
// avoiding the cost of spawning the binary as a subprocess.
type testApp struct {
	*App
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	var stdout, stderr bytes.Buffer
	a := New(":memory:", nil, &stdout, &stderr)
	t.Cleanup(func() { _ = a.Close() })
	return &testApp{App: a, stdout: &stdout, stderr: &stderr}
}

func (ta *testApp) decodeStdout(t *testing.T, v any) {
	t.Helper()
	if err := json.Unmarshal(ta.stdout.Bytes(), v); err != nil {
		t.Fatalf("decode stdout %q: %v", ta.stdout.String(), err)
	}
}

func (ta *testApp) stdoutString() string { return ta.stdout.String() }
func (ta *testApp) stderrString() string { return ta.stderr.String() }

func (ta *testApp) reset() {
	ta.stdout.Reset()
	ta.stderr.Reset()
}

// run is a small convenience for the common 'dispatch one command, decode the
// JSON response into v' pattern.
func (ta *testApp) run(t *testing.T, v any, cmd string, args ...string) {
	t.Helper()
	ta.reset()
	if err := ta.Dispatch(cmd, args); err != nil {
		t.Fatalf("Dispatch %q: %v\nstderr=%q", cmd, err, ta.stderrString())
	}
	if v != nil {
		ta.decodeStdout(t, v)
	}
}

// listResp is the {"entries": [...]} envelope tests decode into. Generic
// over the entry shape (EntrySlim, Entry, EntryWithSnippet) so each test can
// pick the shape it cares about.
type listResp[T any] struct {
	Entries []T `json:"entries"`
}

// runList dispatches a command that returns the entries envelope and returns
// just the slice.
func runList[T any](t *testing.T, ta *testApp, cmd string, args ...string) []T {
	t.Helper()
	var resp listResp[T]
	ta.run(t, &resp, cmd, args...)
	return resp.Entries
}

// --- harness smoke test ---------------------------------------------------

func TestHarness_UnknownCommand(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("nope", nil)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("error = %v, want 'unknown command' substring", err)
	}
}

// When users reach for an ait-shaped verb like `ant note add ...`, dispatch
// should still report it as unknown but nudge toward `ant add --kind note`.
func TestHarness_UnknownKindVerbSuggests(t *testing.T) {
	cases := []struct {
		cmd  string
		kind string
	}{
		{"note", "note"},
		{"notes", "note"},
		{"adr", "adr"},
		{"adrs", "adr"},
		{"pivot", "pivot"},
		{"pivots", "pivot"},
	}
	for _, c := range cases {
		t.Run(c.cmd, func(t *testing.T) {
			ta := newTestApp(t)
			err := ta.Dispatch(c.cmd, []string{"add", "some text"})
			if err == nil {
				t.Fatalf("expected error for unknown command %q", c.cmd)
			}
			msg := err.Error()
			if !strings.Contains(msg, "--kind "+c.kind) {
				t.Errorf("error = %v, want suggestion containing '--kind %s'", err, c.kind)
			}
		})
	}
}

func TestHarness_UnknownNonKindCommand_NoSuggestion(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("frobnicate", nil)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if strings.Contains(err.Error(), "--kind") {
		t.Errorf("error = %v, should not suggest a kind for arbitrary unknown verbs", err)
	}
}
