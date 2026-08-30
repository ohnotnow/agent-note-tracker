package ant

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Version is the build version. Override at build time with:
//
//	go build -ldflags "-X github.com/ohnotnow/agent-note-tracker/internal/ant.Version=v0.1.0"
//
// The default of "dev" makes a development build obvious in the version
// output and disables the GitHub release-check (no point comparing a dev
// build against tagged releases).
var Version = "dev"

// RepoURL is where ant lives on GitHub. Used to build the GitHub API URL
// for the latest-release lookup and to print the "visit /releases/latest"
// hint when a newer version is available. A var, not a const, so a fork
// can override it at link time.
var RepoURL = "https://github.com/ohnotnow/agent-note-tracker"

// Version handles 'ant version'. Prints a plain-text version line and, for
// non-dev builds, performs a best-effort GitHub release-check so the user
// sees a hint when a newer version is available.
//
// Output style mirrors ait — plain text, not JSON. The version command is
// the one place ant breaks from JSON-by-default, because its primary
// audience is a human running it once in a while, not a script.
//
// The release-check is best-effort: a 5-second timeout, errors swallowed
// silently. Offline users see only the local version line.
func (a *App) Version(args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(a.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(a.Stderr, "usage: ant version")
	}
	if err := a.parseFlags(fs, args); err != nil {
		return err
	}

	fmt.Fprintf(a.Stdout, "ant version %s\n", Version)

	if Version == "dev" {
		return nil
	}

	latest, err := checkLatestRelease(RepoURL)
	if err != nil {
		return nil
	}

	if isNewer(latest, Version) {
		fmt.Fprintf(a.Stdout, "A newer version (%s) is available.\n", latest)
		fmt.Fprintf(a.Stdout, "Visit %s/releases/latest to update, or run `ant self-update`.\n", RepoURL)
	} else {
		fmt.Fprintln(a.Stdout, "You are running the latest version.")
	}

	return nil
}

// ghAsset is a single binary attached to a GitHub release.
type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// ghRelease is the slice of the GitHub /releases/latest payload that ant
// cares about — the tag name for version comparison, the markdown body for
// release notes, and the asset list so self-update can find the right binary.
type ghRelease struct {
	TagName string    `json:"tag_name"`
	Body    string    `json:"body"`
	Assets  []ghAsset `json:"assets"`
}

// fetchLatestRelease asks the GitHub API for the latest published release at
// apiURL and returns the parsed payload. The caller supplies the http.Client
// so it can pick an appropriate timeout — 'version' wants a short one (so a
// slow network does not make the command feel hung), while 'self-update'
// needs longer for a multi-megabyte binary download.
func fetchLatestRelease(client *http.Client, apiURL string) (*ghRelease, error) {
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// checkLatestRelease asks the GitHub API for the latest published release
// of repoURL and returns its tag_name. A 5-second timeout keeps a slow or
// unreachable network from making 'ant version' feel hung.
func checkLatestRelease(repoURL string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	rel, err := fetchLatestRelease(client, buildAPIURL(repoURL))
	if err != nil {
		return "", err
	}
	return rel.TagName, nil
}

// buildAPIURL converts a github.com repo URL into the corresponding
// api.github.com /releases/latest endpoint. Tolerates http/https and a
// trailing slash on the input.
func buildAPIURL(repoURL string) string {
	path := strings.TrimPrefix(repoURL, "https://github.com/")
	path = strings.TrimPrefix(path, "http://github.com/")
	path = strings.TrimSuffix(path, "/")
	return "https://api.github.com/repos/" + path + "/releases/latest"
}

// isNewer reports whether latest is a strictly higher semver than current.
// Both inputs are strict three-part semver (vX.Y.Z or X.Y.Z); anything else
// (pre-release suffixes, two-part versions, garbage) returns false. That
// means a non-semver tag never triggers the "newer version available" hint,
// which is the safe default for a best-effort check.
func isNewer(latest, current string) bool {
	parse := func(v string) (int, int, int, bool) {
		v = strings.TrimPrefix(v, "v")
		parts := strings.Split(v, ".")
		if len(parts) != 3 {
			return 0, 0, 0, false
		}
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		patch, err3 := strconv.Atoi(parts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			return 0, 0, 0, false
		}
		return major, minor, patch, true
	}

	lMaj, lMin, lPat, lok := parse(latest)
	cMaj, cMin, cPat, cok := parse(current)
	if !lok || !cok {
		return false
	}

	if lMaj != cMaj {
		return lMaj > cMaj
	}
	if lMin != cMin {
		return lMin > cMin
	}
	return lPat > cPat
}
