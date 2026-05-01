---
name: ant
description: Local-first notebook for *why* — design decisions, alternatives rejected, pivots taken, the conversational nuance that gets lost when an issue closes. Sibling to ait. Use when the user asks to record/note a decision, after a load-bearing choice between options, after a library swap or refactor of direction, or to recall prior decisions in an area.
---

# AIT (`ant`) — Agent Notebook Tool

`ant` is the *why* companion to [ait](../../ait)'s *what*. Where ait tracks
open issues and dependencies, ant captures the durable record of decisions
made, alternatives evaluated, and pivots taken — the context a future
session needs to pick up where the last one left off.

The database lives at `.ant/ant.db`, gitignored by default. Notes are
developer-personal working memory; they can be promoted to project
documentation later via `ant export`.

## When to reach for it

Lean conservative — capture entries when the *why* would be hard to
reconstruct from the code alone:

- **The user asks**: "let's make a note of this", "record this", "ADR this".
- **A load-bearing decision was made**: a choice that constrains future
  work and isn't obvious from reading the code.
- **An evaluation between options ended in a pick**: "we tried X but went
  with Y because…".
- **A pivot or library swap**: moving away from one approach to another.

You don't need to capture every small detail — favour fewer high-value
entries over a stream-of-consciousness log. If in doubt, ask the user.

## Conventional kinds

The schema doesn't enforce these — `kind` is a free string — but starting
with three named conventions keeps the notebook navigable:

- **`adr`** — Architecture Decision Record. A load-bearing choice worth
  recording properly: title, context, options considered, the decision,
  the consequences. Longer-form by convention.
- **`note`** — A captured thought, observation, or short rationale. The
  default and the bulk of entries.
- **`pivot`** — A change of direction: "we tried X for Y reason, it
  didn't pan out, here's why we moved on". Often shorter than an ADR but
  worth distinguishing because pivots tell future-you "don't reconsider this
  unless conditions changed".

Promote a `note` to an `adr` later (`ant edit --kind adr <id>`) if it turns
out to matter.

## Capturing entries

```bash
ant init                                  # one-time, sets up .ant/ant.db
ant init --prefix myproject               # override the inferred prefix

ant add --body "rationale text"           # literal body
ant add --body @path/to/file.md           # body from file
echo "rationale" | ant add                # body from stdin

ant add --title "Choose sqlite" --kind adr --issue ait-AbCdE.2 \
        --body "we picked modernc/sqlite over CGO bindings because…"
```

`--issue` is free-form; if you use ait it'll match an ait id, but Jira /
Linear / GitHub ids work just as well — ant doesn't validate the value.

## Recall — the four read-side moments

These are when ant earns its keep. Each returns slim JSON by default; reach
for `ant show <id>` when you need the full body.

| Moment | Command |
| --- | --- |
| **Session start** — "what did we decide recently?" | `ant recent --limit 5` |
| **Touching an area** — "any prior decisions here?" | `ant search "auth"` |
| **Evaluating a library** — "have we looked at this before?" | `ant search "spatie"` |
| **Working on an issue** — "what's the context?" | `ant for <issue-id>` |

`recent` and `search` include a 200-character body snippet so you can
decide whether to follow up with `show`. Multi-word `search` ANDs the terms
across title and body (case-insensitive).

```bash
ant list                                  # slim JSON, newest first
ant list --long                           # full bodies
ant list --human                          # tabular view
ant list --kind adr --since 2026-04-01    # filtered

ant show <id>                             # one entry, full record
```

## Editing and promotion

```bash
ant edit <id> --title "New title"         # change one column
ant edit <id> --issue ""                  # clear a column (empty string)
echo "new body" | ant edit <id>           # replace body via stdin
ant edit <id> --body @path/to/new.md      # replace body from file

ant export <id>                           # render one entry as markdown
ant export --kind adr                     # render every ADR as markdown
ant export --json                         # JSON instead of markdown
```

`export` is how a personal entry gets promoted to project documentation:
pipe it into a PR description, save it as a doc file, share it as a gist.

## Removing entries

```bash
ant delete <id>
```

`ant delete` is intentionally cautious. If something looks wrong with the
warning it prints, **stop and check with the user** — the cost of a
mistaken delete is high (entries don't soft-delete) and the cost of
double-checking is low. Don't routinely retry a refused delete; surface the
refusal to the user instead.

## Output mode

Default output is JSON. Pipe to `jq` for ad-hoc shaping, or use
`ant list --human` when a table is more useful for a person reading along.

## Database location and `--db`

```bash
ant config                                # show prefix, schema_version, db path
ant --db /path/to/other.db <command>      # operate on a different database
```

`.ant/ant.db` lives at the git root and is added to `.gitignore` by `ant
init`. Each project gets its own notebook.

## Quick mental model

- **Capture is cheap.** Don't ceremony it. A two-sentence `note` beats nothing.
- **Recall is targeted.** Pull the right few entries, not the whole notebook.
- **`ant` is for the *why*; `ait` is for the *what*.** Cross-reference
  with `--issue`.
- **Personal by default.** Promote what deserves promotion; let the rest
  stay personal working memory.
