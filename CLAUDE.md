# CLAUDE.md — agent-note-tracker (`ant`)

Project-specific guidance. The README has the user-facing tour; this
file is the bit you (Claude) need to work effectively in the codebase.

## What this is

A small Go CLI that stores per-project notebook entries in SQLite. It
captures the *why* of a project — design decisions, alternatives
rejected, pivots — and is the sibling of `ait` (which tracks open
work). They're meant to be used together: `--issue` on an `ant` entry
links to whatever issue id the user uses (often an `ait` id).

If you have the `ant` skill loaded, read its `SKILL.md` —
that's the user-facing contract this project implements, and the
mental model the code is shaped around.

## Tooling

- **Go 1.25+** (see `go.mod`). No CGO — pure-Go is deliberate so
  cross-compilation stays clean.
- **SQLite via `modernc.org/sqlite`** (pure Go). Don't swap it for a
  CGO binding without a very good reason.
- **`sqids`** for public ids (`ant-AbCdE` style). Internal DB ids stay
  hidden from users.

Common commands:

```bash
go build -o ant .                # build the binary
go test ./...                    # run the full suite (fast — no subprocesses)
go vet ./...                     # static checks
```

There is no Lando, uv, npm, or similar — this is a single-binary Go
project. Ignore the global CLAUDE.md guidance about Laravel/Lando here.

## Code layout

```
main.go                          CLI entrypoint, --db extraction
internal/ant/app.go              App struct, Dispatch, lazy Store, I/O streams
internal/ant/store.go            DB connection (opens modernc/sqlite)
internal/ant/migrate.go          Forward-only schema migrations
internal/ant/entries.go          Insert/Get/List/Update/Delete entry
internal/ant/prefix.go           Project prefix CRUD
internal/ant/keys.go             sqids public_id generation
internal/ant/paths.go            Project root detection, --db resolution
internal/ant/gitignore.go        Idempotent .gitignore handling
internal/ant/format.go           Markdown rendering
internal/ant/cmd_*.go            One file per command handler (+ _test.go)
internal/ant/version.go          Version + ldflags hook
internal/ant/completion.go       bash/zsh completion scripts
internal/ant/testapp_test.go     Test harness (newTestApp helper)
claude/skills/ant/               The Claude skill for using ant
```

One command per `cmd_*.go` file with a paired `cmd_*_test.go`. New
commands should follow that pattern and be wired into `Dispatch` in
`app.go`.

## Patterns to follow

- **`App` owns I/O.** Commands write through `a.Stdout` / `a.Stderr`
  and read from `a.Stdin`, never `os.Stdout` etc. directly. This is
  what makes the test harness work without subprocesses.
- **Store opens lazily.** `App.Store()` opens and migrates on first
  use. Commands that don't need the DB (like `version`) shouldn't
  trigger it.
- **`requireInitialised()`** is the gate for commands that need a real
  initialised DB — use it instead of opening the store directly.
- **JSON output by default.** Most commands emit slim JSON via
  `a.writeJSON(v)`. Human/markdown variants are opt-in flags.
- **Forward-only migrations.** New schema changes append a numbered
  migration in `migrate.go`; never edit a past one.
- **Public ids are `sqids`-encoded.** Don't expose raw rowids.

## Testing

- `go test ./...` runs everything in-process against `:memory:` SQLite.
- New command tests should use `newTestApp(t)` from
  `testapp_test.go` and dispatch via `ta.run(t, &out, "cmd", args...)`.
  See `cmd_*_test.go` for the established style.
- Don't shell out to the built binary in tests — the harness exists
  precisely to avoid that.

## Things not to do

- Don't add a CGO dependency. Pure-Go cross-compile is load-bearing.
- Don't commit anything under `.ant/` — that directory is the user's
  personal notebook database for *this* project, and is gitignored.
- Don't edit historical migrations. Add a new one.
- Don't bypass the `App` I/O streams by reaching for `os.Stdout`/`Stdin`
  directly in command handlers — it breaks the test harness.
- Don't rename public command names or flag names without the user
  explicitly asking; the skill and any user muscle memory depends on
  them.

## Status

The README states v0.x and the schema is still moving — that's
accurate. Breaking changes are acceptable when justified, but the
user prefers small, considered changes over sweeping refactors. If
you find yourself wanting to restructure something large, surface
the idea before doing it.
