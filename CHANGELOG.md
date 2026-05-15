# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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

[1.3.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.3.0
[0.2.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.2.0
[0.1.1]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.1
[0.1.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.0
