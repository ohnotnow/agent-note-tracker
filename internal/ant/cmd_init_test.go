package ant

import (
	"strings"
	"testing"
)

func TestInit_FreshDB_InferredPrefix(t *testing.T) {
	ta := newTestApp(t)
	var got InitResult
	ta.run(t, &got, "init")

	if got.DB != ":memory:" {
		t.Errorf("db = %q, want :memory:", got.DB)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema_version = %d, want %d", got.SchemaVersion, SchemaVersion)
	}
	if !got.Created {
		t.Error("expected created=true on fresh DB")
	}
	if got.Rekeyed {
		t.Error("expected rekeyed=false on fresh DB")
	}
	if got.Prefix == "" {
		t.Error("expected non-empty inferred prefix")
	}
}

func TestInit_ExplicitPrefix(t *testing.T) {
	ta := newTestApp(t)
	var got InitResult
	ta.run(t, &got, "init", "--prefix", "myproj")
	if got.Prefix != "myproj" {
		t.Errorf("prefix = %q, want %q", got.Prefix, "myproj")
	}
}

func TestInit_PrefixGetsSanitised(t *testing.T) {
	ta := newTestApp(t)
	var got InitResult
	ta.run(t, &got, "init", "--prefix", "My-Cool Project!")
	if got.Prefix != "mycoolproject" {
		t.Errorf("prefix = %q, want %q", got.Prefix, "mycoolproject")
	}
}

func TestInit_RejectsUnusablePrefix(t *testing.T) {
	ta := newTestApp(t)
	err := ta.Dispatch("init", []string{"--prefix", "!!!"})
	if err == nil {
		t.Fatal("expected error for unusable prefix")
	}
	if !strings.Contains(err.Error(), "no usable characters") {
		t.Errorf("error = %v, want substring 'no usable characters'", err)
	}
}

func TestInit_ReinitKeepsExistingPrefix(t *testing.T) {
	ta := newTestApp(t)
	var first InitResult
	ta.run(t, &first, "init", "--prefix", "alpha")

	var second InitResult
	ta.run(t, &second, "init")

	if second.Prefix != "alpha" {
		t.Errorf("second init prefix = %q, want %q", second.Prefix, "alpha")
	}
	if second.Created {
		t.Error("expected created=false on re-init")
	}
	if second.Rekeyed {
		t.Error("expected rekeyed=false when prefix unchanged")
	}
}

func TestInit_ReinitWithDifferentPrefixSignalsRekey(t *testing.T) {
	ta := newTestApp(t)
	var first InitResult
	ta.run(t, &first, "init", "--prefix", "alpha")

	var second InitResult
	ta.run(t, &second, "init", "--prefix", "beta")

	if second.Prefix != "beta" {
		t.Errorf("prefix = %q, want %q", second.Prefix, "beta")
	}
	if !second.Rekeyed {
		t.Error("expected rekeyed=true when prefix changes")
	}
}
