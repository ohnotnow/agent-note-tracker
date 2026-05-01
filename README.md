# ant — Agent Notebook Tool

> **Status: work-in-progress (v0.x).** The CLI and the on-disk schema are
> both still moving. Pin a tagged version if you're going to depend on it.

`ant` is a small local notebook for the **why** of a project. The design
decisions, the alternatives you tried and binned, the awkward pivots, the
bit of context that vanishes the moment an issue gets closed.

It's the sibling of [ait](https://github.com/ohnotnow/agent-issue-tracker).
Where ait keeps track of the *open work* — issues, dependencies, what's
ready to pick up next — `ant` keeps the *paper trail* of how you got here.

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

`ait flush --summary` leaves a tiny breadcrumb when you clear issues out.
That's fine in the moment, but it's not much help when you come back to a
project six months later wondering why on earth you picked *this* library,
or why a refactor went sideways, or why you abandoned an approach that
sounds perfectly reasonable in hindsight.

I tried a free-form `NOTEBOOK.md` for a while. It works, right up until
it doesn't — by the time the notebook actually matters it's thousands of
lines long, and asking an agent to wade through the lot is just noise.

`ant` is the lazy fix for that. Adding an entry is a one-liner, and when
the agent needs context later it can pull the handful of entries that
actually relate to whatever it's working on, rather than re-reading
everything you've ever written down.

## Install

Build from source — Go 1.25+:

```sh
go install github.com/ohnotnow/agent-note-tracker@latest      # once tagged
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

The database lives at `.ant/ant.db` at the git repo root, and `ant init`
quietly adds `.ant/` to `.gitignore` for you.

That's on purpose. A lot of what ends up in here is personal working
memory: preferences, half-formed grumbles, "fine, just ship it" decisions
made at 5pm on a Friday. None of that needs to be a team artefact. When
something does earn its place — an actual ADR, a useful post-mortem —
promote it with `ant export` and pipe the markdown wherever it belongs:
a PR description, a docs file, a gist.

## Conventional kinds

There are three named kinds to start with. The schema doesn't enforce
them, so feel free to invent your own as a project grows:

| Kind | When to use |
| --- | --- |
| `note` (default) | Captured thoughts, observations, short rationales. The bulk of entries. |
| `adr` | Architecture Decision Records — load-bearing choices worth writing down properly. |
| `pivot` | Changes of direction: "we tried X, here's why we moved on". |

If a `note` turns out to matter more than you thought, promote it later
with `ant edit --kind adr <id>`.

## Linking to issue trackers

`--issue` is a free-form string. `ant` won't validate or resolve it, so
use ait ids if you're using ait, or Jira / Linear / GitHub ids otherwise.
`ant for <id>` returns entries with an exact match — nothing fancier.

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

JSON by default, because that's what agents and pipelines want. If you're
the one reading the output, `ant list --human` gives you a tabular view,
`--long` includes the full bodies, and `ant export` produces markdown.

## `ant delete <id> --force`

`ant delete <id>` won't actually delete anything by itself. That's on
purpose — it's far too easy to nuke the wrong entry from a shell loop.
Run it without `--force` and it just prints a warning to stderr saying
what it *would* have deleted, exits non-zero, and leaves the DB alone:

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

The deleted record gets echoed to stdout on the way out, so at least
there's something in your scrollback if you immediately regret it. There's
no soft-delete by design: entries are either there or they're not, and
recovery means your shell history or your backups.

## `--db <path>`

```sh
ant --db /path/to/other.db <command>
```

Handy for scratch databases, testing, or hopping between worktrees. The
special path `:memory:` opens an in-memory SQLite instance, which is what
the test suite uses.

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

MIT — see [LICENSE](LICENSE).
