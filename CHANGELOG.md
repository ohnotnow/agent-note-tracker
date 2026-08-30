# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.7.0] - 2026-08-29
### Fixed
- The self-update redirect for go-install users printed `go install github.com/ohnotnow/agent-note-tracker@latest`, a command that could never work: `go.mod` declared the bare module name `agent-note-tracker`, and Go refuses to install a module whose declared path doesn't match the path you asked for. The module is now declared under its full GitHub path and the hint points at `github.com/ohnotnow/agent-note-tracker/cmd/ant@latest`, which installs a binary actually named `ant`. Mirrors the same fix in the sibling `ait` tool (v1.15.0). `go install` works from this release's tag onwards.

### Changed
- The main package moved from the repository root to `cmd/ant/` in support of the above (with `go install`, the binary is named after the last element of the package path). Local builds are now `go build -o ant ./cmd/ant`; release builds and docs updated to match. No change to the CLI itself.

## [1.6.0] - 2026-07-12
### Added
- `ant init` JSON output gains a `note` field ("no .git directory — not adding .ant/ to .gitignore") when the gitignore step is skipped because the project isn't a git repository yet. Previously the skip was silent, leaving you to discover it at the next `git status`. Mirrors `ait`'s behaviour.

## [1.5.0] - 2026-06-07
### Added
- `ant update` accepted as an alias for `ant edit`, matching the sibling `ait` tool (which takes both). Continues the `create`/`add` aliasing from 1.4.2. Input-only — output, help, and docs still use `edit`.
- `--type` and `--description` accepted as input aliases for `--kind` and `--body` on `add`, `edit`, `list`, and `export`, matching `ait`'s flag names. Output keys are unchanged (still `kind`). Supplying both spellings of the same field at once is rejected.
- `usage` error code in the JSON error-envelope vocabulary, paired with shell exit code `64` (`EX_USAGE`) for command-line grammar mistakes — mirroring `ait`, so a wrapper can read the same signal from both tools.

### Changed
- Command-line grammar failures (unknown command, unknown flag, a missing or extra positional argument, mutually-exclusive flags, a `--db` with no value) now return error code `usage` and exit `64`, rather than `validation_error`/`internal_error` at exit `1`. Genuine value errors (empty body, an unparseable `--since`, an out-of-range `--limit`) still use `validation_error` at exit `1`.

### Fixed
- An invalid or unknown flag no longer dumps Go's raw `flag`-package usage block to stderr alongside the JSON error. Only the clean `{"error": …}` envelope is emitted now — for example `ant list --tree` previously printed a usage block plus an `internal_error`, and now returns just a `usage` envelope.

## [1.4.2] - 2026-05-16
### Added
- `ant create` as an alias for `ant add`. Matches the verb used by the sibling `ait` tool, so agents juggling both no longer trip over the differing command names.

## [1.4.1] - 2026-05-15
### Changed
- `ant --version` and `ant -v` now work as aliases for `ant version`, matching `--help`/`-h` and the convention used by `ait`.
- The "newer version available" hint now ends with "or run `ant self-update`", so the next step is visible inline rather than only in the docs.

## [1.4.0] - 2026-05-15
### Added
- `ant append <id>` command. Joins new content onto an existing entry's body with a blank-line-flanked markdown `---` rule, so the entry still renders as distinct sections when exported through a markdown viewer. Supports `--body <text>`, `--body @<file>`, `--body -` for explicit stdin, and implicit stdin when `--body` is omitted.
- `--body -` recognised by `add`, `edit`, and `append` as an explicit stdin source, in line with common Unix tool convention. The implicit "no `--body`, read stdin" behaviour is unchanged.
- Unknown-command hint when the typed verb matches a conventional kind (e.g. `ant note add ...` → suggests `ant add --kind note ...`). Aimed at the `ait`/`ant` muscle-memory slip.

### Changed
- `add`, `edit`, and `append` `--help` output now includes a single-quoted heredoc example.

## [1.3.0] - 2026-05-04
### Added
- `ant self-update` command. Downloads the matching binary for your OS/arch from the latest GitHub release, verifies it against the published `SHA256SUMS`, and atomically replaces the running executable. Release notes appear inside the confirmation prompt.
- `--check` flag on `self-update` reports availability without installing (exit 0 = current, 1 = newer available, 2 = lookup failed).
- `--yes` flag on `self-update` skips the confirmation prompt for non-interactive use.
- Short-circuits for installs that shouldn't self-update: dev builds refuse, Homebrew and `go install` installs point you at the right tool, and unwriteable install dirs report up front rather than failing partway.

### Changed
- Internal: introduced `ExitWithCode` so commands can request a specific shell exit status, and a "silent exit" path that skips the JSON error envelope when a non-zero exit isn't actually an error condition (used by `self-update --check`).

## [1.2.1] - 2026-05-04
### Changed
- Release workflow now publishes a `SHA256SUMS` file alongside the cross-platform binaries, so downloaders (and `self-update` from v1.3.0) can verify their build.

## [1.2.0] - 2026-05-04
### Added
- Shell completion now suggests entry ids for `show`, `edit`, `delete`, and `export` (bash and zsh). Type `ant show ant-<tab>` and the shell will offer matching ids sourced from `ant list`.
- Zsh completion gained per-flag descriptions for `add`, `edit`, and `export`, and now distinguishes positional id completion from flag-value completion.

## [1.1.0] - 2026-05-03
### Added
- Structured JSON error envelope on stderr (`{"error": {"code", "message"}}`) with a stable code vocabulary (`not_found`, `validation_error`, `conflict`, `confirmation_required`, `uninitialised`, `internal_error`) — agents can branch on `code` instead of string-matching stderr text. Mirrors `ait`'s vocabulary so a single switch works across both tools.
- `--long` flag on `add`, `edit`, and `delete` to return the full entry record. Default output is now the slim projection.

### Changed
- Slim JSON is now the default response for `add`, `edit`, `delete`, `list`, `recent`, `search`, `for`, and `foundation`. Pass `--long` (or use `show`) for the full body.
- List-style commands (`list`, `recent`, `search`, `for`, `export --json`) now wrap results in `{"entries": [...]}` rather than emitting a bare array, leaving room for sibling fields (counts, cursors) without breaking consumers.

## [1.0.0] - 2026-05-03
First stable release. No code changes from `v0.3.0` — this tag marks the public API (commands, flags, JSON shapes) as stable enough to commit to semver going forward.

## [0.3.0] - 2026-05-02

## [0.2.0] - 2026-05-02

## [0.1.1] - 2026-05-01

## [0.1.0] - 2026-05-01

[1.7.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.4.2...v1.5.0
[1.4.2]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.4.1...v1.4.2
[1.4.1]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.4.0...v1.4.1
[1.4.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.3.0
[0.2.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.2.0
[0.1.1]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.1
[0.1.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.0
