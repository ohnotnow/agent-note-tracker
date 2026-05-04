package ant

import (
	"strings"
	"testing"
)

// completionFor dispatches `ant completion <shell>` and returns stdout.
func completionFor(t *testing.T, shell string) string {
	t.Helper()
	ta := newTestApp(t)
	if err := ta.Dispatch("completion", []string{shell}); err != nil {
		t.Fatalf("completion %s: %v\nstderr=%q", shell, err, ta.stderrString())
	}
	return ta.stdoutString()
}

func TestCompletion_BashContainsIDAwareness(t *testing.T) {
	out := completionFor(t, "bash")

	// All commands that take an entry id should be listed in id_commands so
	// the script offers id completion for them.
	for _, cmd := range []string{"show", "edit", "delete", "export"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("bash completion missing %q in id_commands list", cmd)
		}
	}

	// The script must invoke `ant list` to source ids and parse the JSON
	// "id" field out of the slim envelope.
	if !strings.Contains(out, "ant list") {
		t.Error("bash completion does not call `ant list` to source ids")
	}
	if !strings.Contains(out, `"id":`) {
		t.Error("bash completion does not parse the JSON id field")
	}
}

func TestCompletion_ZshContainsIDHelper(t *testing.T) {
	out := completionFor(t, "zsh")

	if !strings.Contains(out, "_ant_entry_ids") {
		t.Error("zsh completion missing _ant_entry_ids helper")
	}
	if !strings.Contains(out, "ant list") {
		t.Error("zsh completion does not call `ant list` to source ids")
	}
	// Each id-aware command should reference the helper.
	for _, cmd := range []string{"show", "edit", "delete", "export"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("zsh completion missing case for %q", cmd)
		}
	}
}

func TestCompletion_UnknownShellErrors(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("completion", []string{"fish"})
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error = %v, want 'unsupported shell' substring", err)
	}
}
