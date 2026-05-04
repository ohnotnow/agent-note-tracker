# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.2.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ohnotnow/agent-note-tracker/compare/v0.3.0...v1.0.0
[0.3.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.3.0
[0.2.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.2.0
[0.1.1]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.1
[0.1.0]: https://github.com/ohnotnow/agent-note-tracker/releases/tag/v0.1.0
