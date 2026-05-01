# ant — Agent Notebook Tool

`ant` is a local-first notebook for the **why** of a project: design
decisions, alternatives evaluated and rejected, pivots taken, the
conversational nuance that disappears when an issue closes.

It is the sibling of [ait](https://github.com/Ohffs/ait): where ait tracks
*open work* (issues, dependencies, what's ready next), `ant` tracks the
*durable record* of decisions made along the way.

```
$ ant add --kind adr --title "Choose sqlite" --issue ait-AbCdE.2 \
          --body "modernc/sqlite over CGO bindings — pure-Go cross-compiles cleanly"

$ ant recent
[
  {
    "id": "ant-AkRXV",
    "kind": "adr",
    "title": "Choose sqlite",
    "issue_id": "ait-AbCdE.2",
    "created_at": "2026-05-01T18:22:22Z",
    "snippet": "modernc/sqlite over CGO bindings — pure-Go cross-compiles cleanly"
  }
]
```

## Why it exists

`ait flush --summary` leaves a small breadcrumb when issues are cleared.
But coming back to a project weeks or months later, those breadcrumbs
aren't enough — you want to know *why* a library was picked, *why* a
refactor went the way it did, *why* an approach was abandoned. A free-form
`NOTEBOOK.md` works for a while but scales badly: by the time it matters,
asking an agent to read thousands of lines is signal-to-noise collapse.

`ant` keeps capture cheap and recall targeted. The agent pulls the right
few entries when it needs them, instead of re-reading the whole notebook.

## Install

Build from source — Go 1.25+:

```sh
go install github.com/Ohffs/ant@latest      # once published
# or, in this checkout:
go build -o ant .
```

## Quick start

```sh
ant init                             # creates .ant/ant.db at the git root
ant init --prefix myproj             # override the inferred prefix

ant add --body "rationale"           # capture
ant add --body @path/to/file.md      # capture from file
echo "rationale" | ant add           # capture from stdin

ant recent                           # 5 most recent, with snippets
ant search "auth refactor"           # multi-term AND across title + body
ant for ait-AbCdE.2                  # entries linked to a specific issue
ant show ant-AkRXV                   # full detail for one entry
ant list --kind adr --human          # filtered table view

ant edit ant-AkRXV --title "New"     # update fields
ant export ant-AkRXV                 # render one entry as markdown
ant export --kind adr                # render every ADR as markdown
ant export --json                    # JSON instead of markdown
```

## Personal by default

The database lives at `.ant/ant.db` at the git repo root. `ant init` adds
`.ant/` to `.gitignore` automatically.

That's deliberate: many entries are personal working memory — preferences,
"just ship it" pressure decisions, throwaway thoughts. They don't all need
to be team artefacts. When something earns it, promote it with `ant
export`: pipe the markdown into a PR description, save it as a docs file,
share it as a gist.

## Conventional kinds

Three named conventions to start with — the schema doesn't enforce them, so
projects can grow their own:

| Kind | When to use |
| --- | --- |
| `note` (default) | Captured thoughts, observations, short rationales. The bulk of entries. |
| `adr` | Architecture Decision Records — load-bearing choices worth recording properly. |
| `pivot` | Changes of direction: "we tried X, here's why we moved on". |

Promote a `note` to an `adr` later with `ant edit --kind adr <id>` if it
turns out to matter.

## Linking to issue trackers

`--issue` is a free-form string — `ant` doesn't validate or resolve it. Use
ait ids if you use ait, Jira/Linear/GitHub ids otherwise. `ant for <id>`
will return entries with an exact match.

## Command reference

| Command | Purpose |
| --- | --- |
| `ant init [--prefix p]` | Create `.ant/`, run migrations, set the prefix, add gitignore entry |
| `ant config` | Show prefix, schema version, resolved DB path |
| `ant add` | Capture a new entry (body via `--body`, `--body @file`, or stdin) |
| `ant show <id>` | Full record for one entry |
| `ant edit <id>` | Update body / title / kind / issue (empty `--title`/`--issue` clears) |
| `ant delete <id>` | Remove an entry (refuses without `--force`, see below) |
| `ant list` | List entries; `--long`, `--human`, filters: `--kind`, `--issue`, `--since` |
| `ant recent [--limit N]` | N most recent entries with body snippets |
| `ant search <query>` | Case-insensitive AND-of-terms across title + body |
| `ant for <issue-id>` | Entries linked to a specific issue id (exact match) |
| `ant export [<id>]` | Render entries as markdown; `--json` for JSON |
| `ant version` | Print the build version |
| `ant completion {bash,zsh}` | Print a shell completion script |

## Output

JSON by default — agent-friendly and pipe-friendly. `ant list --human`
gives a tabular view for humans; `--long` includes full bodies; `ant
export` produces markdown.

## `ant delete <id> --force`

`ant delete <id>` deliberately refuses to act on its own. It prints a
warning to stderr describing what *would* be deleted, exits non-zero, and
leaves the database alone:

```sh
$ ant delete ant-AkRXV
would delete ant-AkRXV — "Choose sqlite" (adr, 2026-05-01T18:22:22Z)
this is irreversible; refusing to act without confirmation.
```

Pass `--force` to actually remove the row:

```sh
$ ant delete ant-AkRXV --force
{ "id": "ant-AkRXV", ... }
```

The deleted record echoes to stdout so the action is visible in
scrollback. Soft-delete is intentionally not implemented — entries are
either present or gone, with no recovery beyond your shell history and
backups.

## `--db <path>`

```sh
ant --db /path/to/other.db <command>
```

Useful for scratch databases, testing, or operating across worktrees. The
special path `:memory:` opens an in-memory SQLite instance (used by the
test suite).

## Build

```sh
go build -o ant .

# with a stamped version
go build -ldflags "-X agent-note-tracker/internal/ant.Version=v0.1.0" -o ant .
```

## Layout

```
main.go                         CLI entrypoint, --db extraction
internal/ant/app.go             App struct, command dispatch
internal/ant/store.go           DB connection
internal/ant/migrate.go         Forward-only schema migrations
internal/ant/entries.go         Insert/Get/List/Update/Delete entry
internal/ant/prefix.go          Project prefix CRUD + Rekey
internal/ant/keys.go            sqids public_id generation
internal/ant/paths.go           Project root detection, path resolution
internal/ant/gitignore.go       Idempotent .gitignore handling
internal/ant/format.go          Markdown rendering
internal/ant/cmd_*.go           One file per command handler
internal/ant/version.go         Version + ldflags hook
internal/ant/completion.go      bash/zsh completion scripts
claude/SKILL.md                 Claude Code skill documentation
```

## License

TBD.
