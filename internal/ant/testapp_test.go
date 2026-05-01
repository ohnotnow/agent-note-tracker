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
