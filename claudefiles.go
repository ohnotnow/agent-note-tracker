// Package antfiles carries the Claude Code skill that goes with this version
// of ant, built into the binary so `ant install-skills` can put it in place
// and `ant version` can tell when the installed copy differs. It lives at
// the repo root because go:embed only reaches files at or below the
// embedding package's directory.
package antfiles

import "embed"

//go:embed claude
var Claude embed.FS
