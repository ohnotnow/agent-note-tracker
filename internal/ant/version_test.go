package ant

import (
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v1.0.0", "v0.9.0", true},
		{"v0.2.0", "v0.1.0", true},
		{"v0.1.1", "v0.1.0", true},
		{"v1.0.0", "v1.0.0", false},
		{"v0.1.0", "v0.2.0", false},
		{"v0.1.0", "v0.1.1", false},
		{"v1.10.0", "v1.9.0", true},
		{"v1.9.0", "v1.10.0", false},
		{"invalid", "v1.0.0", false},
		{"v1.0.0", "invalid", false},
		{"", "", false},
		{"v1.0", "v1.0.0", false},
		// Pre-release tags fall through to false — strict semver only.
		{"v1.0.0-rc1", "v0.9.0", false},
		// 'v' prefix is optional and mixing is fine.
		{"1.2.3", "v1.2.2", true},
	}

	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			got := isNewer(tt.latest, tt.current)
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}

func TestBuildAPIURL(t *testing.T) {
	tests := []struct {
		repoURL string
		want    string
	}{
		{
			"https://github.com/ohnotnow/agent-note-tracker",
			"https://api.github.com/repos/ohnotnow/agent-note-tracker/releases/latest",
		},
		{
			"https://github.com/ohnotnow/agent-note-tracker/",
			"https://api.github.com/repos/ohnotnow/agent-note-tracker/releases/latest",
		},
		{
			"http://github.com/someuser/somefork",
			"https://api.github.com/repos/someuser/somefork/releases/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.repoURL, func(t *testing.T) {
			got := buildAPIURL(tt.repoURL)
			if got != tt.want {
				t.Errorf("buildAPIURL(%q) = %q, want %q", tt.repoURL, got, tt.want)
			}
		})
	}
}

// Dev builds short-circuit before the network check, so this is testable
// without touching GitHub. It also confirms the new plain-text contract.
func TestVersion_DevSkipsNetwork(t *testing.T) {
	ta := newTestApp(t)
	if err := ta.Dispatch("version", nil); err != nil {
		t.Fatalf("version: %v", err)
	}
	out := ta.stdoutString()
	if !strings.HasPrefix(out, "ant version dev") {
		t.Errorf("stdout = %q, want it to start with 'ant version dev'", out)
	}
	if strings.Contains(out, "newer version") || strings.Contains(out, "latest version") {
		t.Errorf("dev build should not emit release-check output, got %q", out)
	}
}
