package ant

import (
	"sort"
	"strings"
	"testing"
)

// insertTestEntry inserts a row with a public_id derived from the given
// prefix and explicit integer id. Used to set up Rekey tests before Add lands.
func insertTestEntry(t *testing.T, s *Store, id int64, prefix string) {
	t.Helper()
	pubID, err := PublicID(prefix, id)
	if err != nil {
		t.Fatalf("PublicID: %v", err)
	}
	_, err = s.DB.Exec(
		`INSERT INTO entries (id, public_id, kind, body, created_at, updated_at)
		 VALUES (?, ?, 'note', 'body', '2026-05-01T00:00:00Z', '2026-05-01T00:00:00Z')`,
		id, pubID,
	)
	if err != nil {
		t.Fatalf("insert entry %d: %v", id, err)
	}
}

func TestRekey_UpdatesPrefixAndPublicIDs(t *testing.T) {
	s := openMemoryStore(t)
	if err := s.Migrate(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPrefix("alpha"); err != nil {
		t.Fatal(err)
	}
	insertTestEntry(t, s, 1, "alpha")
	insertTestEntry(t, s, 2, "alpha")
	insertTestEntry(t, s, 3, "alpha")

	if err := s.Rekey("beta"); err != nil {
		t.Fatalf("Rekey: %v", err)
	}

	prefix, _, err := s.GetPrefix()
	if err != nil {
		t.Fatal(err)
	}
	if prefix != "beta" {
		t.Errorf("prefix = %q, want beta", prefix)
	}

	rows, err := s.DB.Query(`SELECT id, public_id FROM entries ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type row struct {
		id  int64
		pub string
	}
	var got []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.pub); err != nil {
			t.Fatal(err)
		}
		got = append(got, r)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(got))
	}

	for _, r := range got {
		if !strings.HasPrefix(r.pub, "beta-") {
			t.Errorf("id %d: public_id = %q, want beta- prefix", r.id, r.pub)
		}
	}

	// The sqid suffix is derived from the integer id, so re-keying back to
	// the original prefix should restore the original public_ids.
	if err := s.Rekey("alpha"); err != nil {
		t.Fatal(err)
	}

	rows2, _ := s.DB.Query(`SELECT public_id FROM entries ORDER BY id`)
	var restored []string
	for rows2.Next() {
		var p string
		_ = rows2.Scan(&p)
		restored = append(restored, p)
	}
	rows2.Close()

	want := []string{}
	for _, id := range []int64{1, 2, 3} {
		p, _ := PublicID("alpha", id)
		want = append(want, p)
	}
	sort.Strings(restored)
	sort.Strings(want)
	for i := range want {
		if restored[i] != want[i] {
			t.Errorf("restored[%d] = %q, want %q", i, restored[i], want[i])
		}
	}
}

func TestRekey_NoEntriesIsFine(t *testing.T) {
	s := openMemoryStore(t)
	if err := s.Migrate(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetPrefix("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := s.Rekey("beta"); err != nil {
		t.Fatalf("Rekey on empty entries: %v", err)
	}
	prefix, _, err := s.GetPrefix()
	if err != nil {
		t.Fatal(err)
	}
	if prefix != "beta" {
		t.Errorf("prefix = %q, want beta", prefix)
	}
}

func TestInit_RekeyChangesEntryIDs(t *testing.T) {
	ta := newTestApp(t)
	ta.run(t, nil, "init", "--prefix", "alpha")

	store, err := ta.Store()
	if err != nil {
		t.Fatal(err)
	}
	insertTestEntry(t, store, 42, "alpha")

	var got InitResult
	ta.run(t, &got, "init", "--prefix", "beta")
	if !got.Rekeyed {
		t.Fatal("expected rekeyed=true")
	}

	var pubID string
	if err := store.DB.QueryRow(`SELECT public_id FROM entries WHERE id = 42`).Scan(&pubID); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(pubID, "beta-") {
		t.Errorf("public_id after rekey = %q, want beta- prefix", pubID)
	}
}
