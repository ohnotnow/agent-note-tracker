package ant

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

var testSkills = fstest.MapFS{
	"skills/ant/SKILL.md":               {Data: []byte("skill v2\n")},
	"skills/ant/BACKFILL_FOUNDATION.md": {Data: []byte("backfill v2\n")},
}

func install(t *testing.T, cfg installConfig) (string, error) {
	t.Helper()
	var out bytes.Buffer
	if cfg.files == nil {
		cfg.files = testSkills
	}
	if cfg.stdin == nil {
		cfg.stdin = strings.NewReader("")
	}
	cfg.stdout = &out
	err := installSkills(cfg)
	return out.String(), err
}

func readOr(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return "<missing>"
	}
	return string(b)
}

func TestInstallSkills(t *testing.T) {
	t.Run("the real build carries the ant skill", func(t *testing.T) {
		list, err := skillFilesFor(skillFiles, "x")
		if err != nil {
			t.Fatal(err)
		}
		var rels []string
		for _, f := range list {
			rels = append(rels, f.rel)
		}
		want := []string{"skills/ant/BACKFILL_FOUNDATION.md", "skills/ant/SKILL.md"}
		if strings.Join(rels, " ") != strings.Join(want, " ") {
			t.Fatalf("embedded = %v, want %v", rels, want)
		}
	})

	t.Run("missing files are added and identical ones left alone", func(t *testing.T) {
		dir := t.TempDir()
		out, err := install(t, installConfig{claudeDir: dir})
		if err != nil || readOr(t, filepath.Join(dir, "skills/ant/SKILL.md")) != "skill v2\n" {
			t.Fatalf("err %v, out:\n%s", err, out)
		}
		out, _ = install(t, installConfig{claudeDir: dir})
		if strings.Count(out, "unchanged") != 2 {
			t.Fatalf("second run:\n%s", out)
		}
	})

	t.Run("with no terminal, a changed file is left as it is and reported", func(t *testing.T) {
		dir := t.TempDir()
		install(t, installConfig{claudeDir: dir})
		mine := filepath.Join(dir, "skills/ant/SKILL.md")
		os.WriteFile(mine, []byte("my tuned skill\n"), 0o644)
		out, err := install(t, installConfig{claudeDir: dir})
		if err != nil || readOr(t, mine) != "my tuned skill\n" || !strings.Contains(out, "skills/ant/SKILL.md: differs, left as it is") {
			t.Fatalf("err %v, file %q, out:\n%s", err, readOr(t, mine), out)
		}
	})

	t.Run("at a terminal: N keeps it, d shows the diff and asks again, y replaces it", func(t *testing.T) {
		dir := t.TempDir()
		install(t, installConfig{claudeDir: dir})
		mine := filepath.Join(dir, "skills/ant/SKILL.md")
		os.WriteFile(mine, []byte("my tuned skill\n"), 0o644)

		out, _ := install(t, installConfig{claudeDir: dir, interactive: true, stdin: strings.NewReader("\n")})
		if readOr(t, mine) != "my tuned skill\n" || !strings.Contains(out, "kept your copy") {
			t.Fatalf("N: file %q, out:\n%s", readOr(t, mine), out)
		}
		out, _ = install(t, installConfig{claudeDir: dir, interactive: true, stdin: strings.NewReader("d\ny\n")})
		if readOr(t, mine) != "skill v2\n" || !strings.Contains(out, "-my tuned skill") || !strings.Contains(out, "+skill v2") || strings.Count(out, "[y/N/d]") != 2 {
			t.Fatalf("d then y: file %q, out:\n%s", readOr(t, mine), out)
		}
	})

	t.Run("input ending mid-prompt stops, and never counts as an answer", func(t *testing.T) {
		dir := t.TempDir()
		install(t, installConfig{claudeDir: dir})
		for _, f := range []string{"skills/ant/SKILL.md", "skills/ant/BACKFILL_FOUNDATION.md"} {
			os.WriteFile(filepath.Join(dir, f), []byte("mine\n"), 0o644)
		}
		_, err := install(t, installConfig{claudeDir: dir, interactive: true, stdin: strings.NewReader("")})
		if err != errAborted || readOr(t, filepath.Join(dir, "skills/ant/SKILL.md")) != "mine\n" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("--force replaces without asking, and --dry-run changes nothing", func(t *testing.T) {
		dir := t.TempDir()
		install(t, installConfig{claudeDir: dir})
		mine := filepath.Join(dir, "skills/ant/SKILL.md")
		os.WriteFile(mine, []byte("mine\n"), 0o644)
		out, _ := install(t, installConfig{claudeDir: dir, dryRun: true})
		if readOr(t, mine) != "mine\n" || !strings.Contains(out, "differs, would ask") {
			t.Fatalf("dry run: out:\n%s", out)
		}
		install(t, installConfig{claudeDir: dir, force: true})
		if readOr(t, mine) != "skill v2\n" {
			t.Fatalf("force: file %q", readOr(t, mine))
		}
	})

	t.Run("version's skills line says whether the installed copies match", func(t *testing.T) {
		dir := t.TempDir()
		if s := skillsStatus(testSkills, dir); !strings.Contains(s, "not installed") {
			t.Fatal(s)
		}
		install(t, installConfig{claudeDir: dir})
		if s := skillsStatus(testSkills, dir); !strings.Contains(s, "up to date") {
			t.Fatal(s)
		}
		os.WriteFile(filepath.Join(dir, "skills/ant/SKILL.md"), []byte("old\n"), 0o644)
		if s := skillsStatus(testSkills, dir); !strings.Contains(s, "skills/ant/SKILL.md") || !strings.Contains(s, "ant install-skills") {
			t.Fatal(s)
		}
	})

	t.Run("the command installs into --dir and needs no database", func(t *testing.T) {
		dir := t.TempDir()
		ta := newTestApp(t)
		if err := ta.Dispatch("install-skills", []string{"--dir", dir}); err != nil {
			t.Fatalf("install-skills: %v\nstderr=%s", err, ta.stderrString())
		}
		if readOr(t, filepath.Join(dir, "skills/ant/SKILL.md")) == "<missing>" {
			t.Fatalf("skill not installed; out:\n%s", ta.stdoutString())
		}
		if ta.store != nil {
			t.Error("install-skills opened the database")
		}
	})
}
