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

## At the start of every session

If the project has an `ant` database, run `ant foundation` early. It
returns the project's single load-bearing vision entry — what the
project is, what it isn't, the spirit you're meant to be working in.
Read it before you start making judgement calls about design, wording,
or trade-offs.

If `ant foundation` reports no foundation recorded *and* the project
shows signs of being established (a populated `CLAUDE.md`, a real
`README.md`, non-trivial git history), offer to backfill one with the
user. If they agree, read
[BACKFILL_FOUNDATION.md](./BACKFILL_FOUNDATION.md) and follow the Q&A
flow there. If they'd rather not — or the project is brand new and
hasn't had its ideation discussion yet — drop it and carry on. Most
projects won't have a foundation and that's fine.

After a project ideation/vision discussion that produced a shared sense
of what the project is and isn't, offer to capture it as the foundation
entry: `ant add --kind foundation --body @path/to/notes.md`. There can
only be one — `add` will refuse a second, and you revise the existing
one with `ant edit` instead.

## When to reach for it

Lean conservative — capture entries when the *why* would be hard to
reconstruct from the code alone:

- **The user asks**: "let's make a note of this", "record this", "ADR this".
- **A load-bearing decision was made**: a choice that constrains future
  work and isn't obvious from reading the code.
- **An evaluation between options ended in a pick**: "we tried X but went
  with Y because…".
- **A pivot or library swap**: moving away from one approach to another.
- **A project ideation phase wrapped up**: the moment after you've
  agreed on what the project is, before any code is written, is the
  right time to capture a `foundation` entry.

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
- **`foundation`** — The project's single vision document (one per
  project, enforced). What this project is and isn't, in the spirit of
  a design doc agreed before code was written. Read it at session
  start; capture one only after a clear ideation/vision discussion.

Promote a `note` to an `adr` later (`ant edit --kind adr <id>`) if it turns
out to matter.

## Capturing entries

```bash
ant init                                  # one-time, sets up .ant/ant.db
ant init --prefix myproject               # override the inferred prefix

ant add --body "rationale text"           # literal body
ant add --body @path/to/file.md           # body from file
echo "rationale" | ant add                # implicit stdin
ant add --human --title "Why X"           # body written in $EDITOR (for a person)
ant add --body - <<'EOF'                  # explicit stdin via heredoc
Multi-line body. Single-quoted 'EOF' means no shell expansion —
backticks and $vars stay literal.
EOF

ant add --title "Choose sqlite" --kind adr --issue ait-AbCdE.2 \
        --body "we picked modernc/sqlite over CGO bindings because…"
```

`--issue` is free-form; if you use ait it'll match an ait id, but Jira /
Linear / GitHub ids work just as well — ant doesn't validate the value.

For non-trivial bodies, the safest options are `--body @file` or a
single-quoted heredoc (`<<'EOF' ... EOF`). The single-quoted delimiter
disables shell expansion, so backticks, `$variables`, and apostrophes
all survive untouched.

### When a new entry overturns an older one

If what you are writing replaces any part of an earlier entry, say so
with `--supersedes`. Readers of the older entry then see a warning
before they act on it. Nobody goes back and edits old entries, so this
is the only way they find out.

```bash
ant add --kind adr --title "Grouping, second wave" \
        --supersedes x-Ed6UZ --supersedes-reason "a story is no longer one folder" \
        --body @path/to/notes.md
```

- Use it whenever any part of the older decision no longer holds, even a
  small part. When it is only part, the reason says which part; leave the
  reason out only when the whole entry is replaced.
- Building on an entry without overturning anything is not superseding.
  Cite it in the body instead ("builds on x-AbCdE, which chose...");
  `show` picks that up as a mention.
- Forgot, or found out later? `ant edit <newer-id> --supersedes <older-id>`
  adds it after the fact, and `--unsupersede <older-id>` removes a wrong
  one. Neither needs a prior `show`, because the body is untouched.

## Recall — the four read-side moments

These are when ant earns its keep. Each returns slim JSON by default; reach
for `ant show <id>` when you need the full body.

| Moment | Command |
| --- | --- |
| **Session start** — "what does this project care about?" | `ant foundation` |
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

**Superseded entries.** If `show` starts with `superseded_by`, or a row
from `list`, `recent`, `search` or `for` carries a `superseded_by`
marker, a newer entry has overturned at least part of this one. Read the
newer entry (and its `reason`, which says which part) before you act on
the older one. A marker on a row is also a good reason to skip straight
to the newer entry. `show` lists the other direction as `supersedes`.

`show` also lists the entries connected to this one by id, worked out
from the bodies on every read:

- `mentions`: entries this body cites, each with its title. If a title
  doesn't fit the sentence that cites it, the id in the body is probably
  mistyped.
- `mentioned_by`: entries whose bodies cite this one. Newer entries that
  build on or revise this one show up here, so check them before acting
  on an older decision.
- `not_in_notebook`: ids with this project's prefix that match no entry.
  Often an ait issue (projects usually share one prefix), sometimes a
  typo; the surrounding sentence usually tells you which.

When asked to read one entry cold ("read ant note x-AbCdE"), look at
`superseded_by` and `mentioned_by` too: the conclusion is in this entry,
but how it was later revised is in those.

## Editing and promotion

```bash
ant edit <id> --title "New title"         # change one column
ant edit <id> --issue ""                  # clear a column (empty string)
echo "new body" | ant edit <id>           # replace body via stdin
ant edit <id> --body @path/to/new.md      # replace body from file
ant edit <id> --human                     # rewrite body in $EDITOR (for a person)
ant edit <id> --body - <<'EOF'            # explicit stdin via heredoc
Rewritten body, apostrophes and `backticks` fine.
EOF
```

**Read before you write.** `ant show <id>` first, then `edit`. The tool records which session showed an entry, and `edit --body/--title/--human` is refused (`unread` error) unless you showed it within the last hour. This is the same rule as reading a file before editing it, and it exists because an agent once overwrote a hand-written entry without having looked at it. If the refusal says the show was too long ago, run `show` again and re-read it; your context may have been compacted since. `--dangerously-skip-read-check` overrides the guard and should almost never appear in your commands.

Two more refusals: an empty body (including an empty heredoc on `--body -`) is rejected, and a replacement under half the length of the existing body is rejected as a likely diff or summary unless you pass `--force`. Only use `--force` when you have just read the old body and mean to replace it. To add to an entry, use `ant append`, which is never guarded.

```bash
ant export <id>                           # render one entry as markdown
ant export --kind adr                     # render every ADR as markdown
ant export --json                         # JSON instead of markdown
```

`edit` replaces the body wholesale. To **grow** an entry instead — a
later clarification, a link to a follow-up note, a dated update — use
`append`:

```bash
ant append <id> --body "2026-05-15: linked to demo-FgHiJ"
ant append <id> --body @path/to/update.md
ant append <id> --body - <<'EOF'
Multi-line update via heredoc.
EOF
```

`append` joins the new content onto the existing body with a blank line,
a markdown `---` rule, and another blank line — so the entry renders as
distinct sections when exported through a markdown viewer. Reach for it
for small clarifications and dated updates; create a sibling `pivot`
entry when the change is structural enough to deserve its own id.

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

### Response shapes

- **Single records** (`add`, `edit`, `delete`, `show`, `foundation`,
  `init`, `config`) return one JSON object on stdout.
- **List commands** (`list`, `recent`, `search`, `for`, `export --json`)
  return `{"entries": [...]}` — always wrapped, never a bare array. The
  envelope leaves room for sibling fields (counts, cursors) without
  breaking consumers, so check `.entries` rather than treating the top
  level as the array itself.
- **Mutations are slim by default.** `add`, `edit`, `append`, and
  `delete` echo back `{id, kind, title?, issue_id?, created_at, superseded_by?}` — no
  body, no `updated_at`. Pass `--long` to get the full record back when
  you actually need it (e.g. confirming a body change took effect).
- **Errors** go to stderr as `{"error": {"code": "...", "message":
  "..."}}` and the process exits non-zero. Stable codes:
  `not_found`, `validation_error`, `conflict`, `confirmation_required`,
  `uninitialised`, `internal_error`. Branch on the code, not the
  message text.

```bash
ant show fake-NOPE 2>&1 >/dev/null | jq -r .error.code   # → not_found
```

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
