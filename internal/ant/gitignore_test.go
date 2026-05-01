package ant

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func withGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	return dir
}

func TestEnsureGitignore_NoFileYet(t *testing.T) {
	dir := withGitDir(t)
	changed, err := EnsureGitignore(dir)
	if err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if !changed {
		t.Error("expected changed=true when creating .gitignore")
	}
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read gitignore: %v", err)
	}
	if string(content) != ".ant/\n" {
		t.Errorf("gitignore content = %q, want %q", content, ".ant/\n")
	}
}

func TestEnsureGitignore_AppendsToExisting(t *testing.T) {
	dir := withGitDir(t)
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("vendor/\nlogs/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	changed, err := EnsureGitignore(dir)
	if err != nil || !changed {
		t.Fatalf("EnsureGitignore changed=%v err=%v", changed, err)
	}
	got, _ := os.ReadFile(path)
	want := "vendor/\nlogs/\n.ant/\n"
	if string(got) != want {
		t.Errorf("gitignore = %q, want %q", got, want)
	}
}

func TestEnsureGitignore_HandlesMissingTrailingNewline(t *testing.T) {
	dir := withGitDir(t)
	path := filepath.Join(dir, ".gitignore")
	_ = os.WriteFile(path, []byte("vendor/"), 0o644)
	changed, err := EnsureGitignore(dir)
	if err != nil || !changed {
		t.Fatalf("EnsureGitignore: changed=%v err=%v", changed, err)
	}
	got, _ := os.ReadFile(path)
	want := "vendor/\n.ant/\n"
	if string(got) != want {
		t.Errorf("gitignore = %q, want %q", got, want)
	}
}

func TestEnsureGitignore_AlreadyListed(t *testing.T) {
	cases := []string{
		"vendor/\n.ant/\n",
		"vendor/\n.ant\n",
		"vendor/\n/.ant/\n",
		"vendor/\n.ant/*\n",
		"# comment\n.ant/\n",
	}
	for _, original := range cases {
		t.Run(original, func(t *testing.T) {
			dir := withGitDir(t)
			path := filepath.Join(dir, ".gitignore")
			_ = os.WriteFile(path, []byte(original), 0o644)
			changed, err := EnsureGitignore(dir)
			if err != nil {
				t.Fatalf("EnsureGitignore: %v", err)
			}
			if changed {
				t.Errorf("expected changed=false for %q", original)
			}
			got, _ := os.ReadFile(path)
			if string(got) != original {
				t.Errorf("file modified: got %q, want %q", got, original)
			}
		})
	}
}

func TestEnsureGitignore_NoGitRepo(t *testing.T) {
	dir := t.TempDir()
	changed, err := EnsureGitignore(dir)
	if err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if changed {
		t.Error("expected changed=false outside a git repo")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !errors.Is(err, fs.ErrNotExist) {
		t.Error("expected no .gitignore created outside a git repo")
	}
}

func TestEnsureGitignore_IgnoresCommentsAndNegations(t *testing.T) {
	dir := withGitDir(t)
	path := filepath.Join(dir, ".gitignore")
	_ = os.WriteFile(path, []byte("# .ant/\n!.ant\n"), 0o644)
	changed, err := EnsureGitignore(dir)
	if err != nil || !changed {
		t.Fatalf("EnsureGitignore: changed=%v err=%v", changed, err)
	}
	got, _ := os.ReadFile(path)
	want := "# .ant/\n!.ant\n.ant/\n"
	if string(got) != want {
		t.Errorf("gitignore = %q, want %q", got, want)
	}
}
