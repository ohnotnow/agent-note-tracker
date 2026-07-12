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

ant edit ant-AkRXV --title "New"     # replace fields wholesale
ant append ant-AkRXV --body "later"  # grow an entry with a dated update
ant export ant-AkRXV                 # render one entry as markdown
ant export --kind adr                # render every ADR as markdown
ant export --json                    # JSON instead of markdown
```

## Personal by default

The database lives at `.ant/ant.db` at the git repo root, and `ant init`
adds `.ant/` to `.gitignore` for you. When it can't — no `.git` directory
yet — its JSON output says so in a `note` field rather than leaving you to
find out at the next `git status`.

That's on purpose. A lot of what ends up in here is personal working
memory: preferences, half-formed grumbles, "fine, just ship it" decisions
made at 5pm on a Friday. None of that needs to be a team artefact. When
something does earn its place — an actual ADR, a useful post-mortem —
promote it with `ant export` and pipe the markdown wherever it belongs:
a PR description, a docs file, a gist.

## Conventional kinds

There are four named kinds to start with. The schema doesn't enforce
them, so feel free to invent your own as a project grows:

| Kind | When to use |
| --- | --- |
| `note` (default) | Captured thoughts, observations, short rationales. The bulk of entries. |
| `adr` | Architecture Decision Records — load-bearing choices worth writing down properly. |
| `pivot` | Changes of direction: "we tried X, here's why we moved on". |
| `foundation` | The single core vision document for the project. Singleton — one per project. |

If a `note` turns out to matter more than you thought, promote it later
with `ant edit --kind adr <id>`.

## Project foundation

The `foundation` kind is special. It's the one entry that captures the
*essence* of a project — what it is, what it isn't, the load-bearing
ideas an agent (or future-you) needs to read before making judgement
calls about design, wording, and trade-offs.

A project has at most one. `ant add --kind foundation` refuses if one
already exists; revise it with `ant edit` instead.

```sh
$ ant add --kind foundation --title "ant — vision" --body @ANT_PLAN.md

$ ant foundation
{
  "id": "ant-AkRXV",
  "kind": "foundation",
  "title": "ant — vision",
  "body": "...the full text, untruncated..."
}
```

`ant foundation` returns the full body (no snippet) — this is the one
read path where truncation would defeat the goal. If no foundation has
been recorded, it exits non-zero with a hint.

The DB is gitignored, so foundation revisions are lossy. If you want a
history of how the vision evolved, render it to markdown periodically
(`ant export <id>`) and commit that file alongside the code.

## Linking to issue trackers

`--issue` is a free-form string. `ant` won't validate or resolve it, so
use ait ids if you're using ait, or Jira / Linear / GitHub ids otherwise.
`ant for <id>` returns entries with an exact match — nothing fancier.

## Command reference

| Command | Purpose |
| --- | --- |
| `ant init [--prefix p]` | Create `.ant/`, run migrations, set the prefix, add gitignore entry |
| `ant config` | Show prefix, schema version, resolved DB path |
| `ant add` (alias: `ant create`) | Capture a new entry (body via `--body`, `--body @file`, or stdin) |
| `ant show <id>` | Full record for one entry |
| `ant edit <id>` | Update body / title / kind / issue (empty `--title`/`--issue` clears) |
| `ant append <id>` | Add content to an existing entry, separated by a markdown `---` rule |
| `ant delete <id>` | Remove an entry (refuses without `--force`, see below) |
| `ant list` | List entries; `--long`, `--human`, filters: `--kind`, `--issue`, `--since` |
| `ant recent [--limit N]` | N most recent entries with body snippets |
| `ant search <query>` | Case-insensitive AND-of-terms across title + body |
| `ant for <issue-id>` | Entries linked to a specific issue id (exact match) |
| `ant foundation` | Print the project's foundation entry, if one has been recorded |
| `ant export [<id>]` | Render entries as markdown; `--json` for JSON |
| `ant version` | Print the build version (and check GitHub for newer releases) |
| `ant self-update` | Update to the latest release; `--check` reports without installing, `--yes` skips the prompt |
| `ant completion {bash,zsh}` | Print a shell completion script |

## Output

JSON by default, because that's what agents and pipelines want. If you're
the one reading the output, `ant list --human` gives you a tabular view,
`--long` includes the full bodies, and `ant export` produces markdown.

## ant delete sneakyness

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

A stamped (non-`dev`) build will, on `ant version`, do a best-effort
GitHub check against the latest tagged release and let you know if a
newer one is available. The check has a 5-second timeout and fails
silently — offline use is unaffected.

## Self-update

```sh
ant self-update            # show release notes, prompt y/N, replace the binary
ant self-update --check    # exit 0 if current, 1 if newer is available, 2 on lookup failure
ant self-update --yes      # skip the prompt
```

`self-update` downloads the matching binary for your OS and architecture
from the latest GitHub release, verifies it against the published
`SHA256SUMS`, and atomically replaces the running executable. Release
notes appear inside the confirmation prompt so there's nothing to
chase down separately.

A few cases short-circuit before any download:

- **Dev builds** refuse — there's no version to compare against, and
  overwriting a hand-built binary with a release one is rarely what you
  want.
- **Homebrew** and **`go install`** installs print a hint pointing you
  at the right tool (`brew upgrade ant`, `go install …@latest`) instead
  of sidestepping your package manager.
- **Unwriteable install dirs** report immediately with a "re-run with
  sudo, or visit /releases/latest" message rather than failing partway.

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
