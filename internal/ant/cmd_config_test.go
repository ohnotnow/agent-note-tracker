package ant

import (
	"strings"
	"testing"
)

func TestConfig_AfterInit(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, nil, "init", "--prefix", "demo")

	var got ConfigResult
	ta.run(t, &got, "config")

	if got.Prefix != "demo" {
		t.Errorf("prefix = %q, want %q", got.Prefix, "demo")
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, SchemaVersion)
	}
	if got.DB != ":memory:" {
		t.Errorf("db = %q, want :memory:", got.DB)
	}
}

func TestConfig_BeforeInit_Errors(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("config", nil)
	if err == nil {
		t.Fatal("expected error when running config without init")
	}
	if !strings.Contains(err.Error(), "ant init") {
		t.Errorf("error = %v, want hint to run 'ant init'", err)
	}
}
