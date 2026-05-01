package ant

import (
	"strings"
	"testing"
)

func TestPublicID_BasicShape(t *testing.T) {
	id, err := PublicID("ant", 1)
	if err != nil {
		t.Fatalf("PublicID: %v", err)
	}
	if !strings.HasPrefix(id, "ant-") {
		t.Errorf("id %q missing 'ant-' prefix", id)
	}
	suffix := strings.TrimPrefix(id, "ant-")
	if len(suffix) < idMinLength {
		t.Errorf("suffix %q shorter than min length %d", suffix, idMinLength)
	}
}

func TestPublicID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := int64(1); i <= 1000; i++ {
		id, err := PublicID("ant", i)
		if err != nil {
			t.Fatalf("PublicID(%d): %v", i, err)
		}
		if seen[id] {
			t.Fatalf("duplicate id %q at i=%d", id, i)
		}
		seen[id] = true
	}
}

func TestPublicID_Deterministic(t *testing.T) {
	a, _ := PublicID("ant", 42)
	b, _ := PublicID("ant", 42)
	if a != b {
		t.Errorf("PublicID not deterministic: %q vs %q", a, b)
	}
}

func TestPublicID_PrefixAffectsOnlyPrefix(t *testing.T) {
	a, _ := PublicID("ant", 7)
	b, _ := PublicID("foo", 7)
	aTail := strings.TrimPrefix(a, "ant-")
	bTail := strings.TrimPrefix(b, "foo-")
	if aTail != bTail {
		t.Errorf("sqid suffix should be prefix-independent; got %q and %q", aTail, bTail)
	}
}

func TestPublicID_RejectsBadInputs(t *testing.T) {
	if _, err := PublicID("", 1); err == nil {
		t.Error("expected error on empty prefix")
	}
	if _, err := PublicID("ant", -1); err == nil {
		t.Error("expected error on negative id")
	}
}
