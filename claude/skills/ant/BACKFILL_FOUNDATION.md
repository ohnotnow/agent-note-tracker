# Backfilling a foundation entry

Referenced from [SKILL.md](./SKILL.md) when a project has no
`foundation` entry but looks established (existing `CLAUDE.md`, real
`README.md`, non-trivial git history) and the user has agreed to fill
one in. If they declined, you shouldn't be reading this — drop it and
carry on.

## Why bother when CLAUDE.md already exists

`CLAUDE.md` tells an agent *how* to work in the codebase: conventions,
tooling, the don't-touch list. A `foundation` entry says what the
project *is for* — in the user's own words, framed for a human rather
than a build system. The two complement each other rather than
overlap: `CLAUDE.md` answers "how should I work here?", foundation
answers "what am I trying to make this thing do, and for whom?".

If part-way through this the user says "honestly, `CLAUDE.md` is
enough" — believe them and stop. This is a genuine 1-in-N case, not
something every project needs.

## The Q&A

Ask one question at a time, conversationally. Wait for the answer
before moving on. Adapt to what the user gives you — if one answer
covers two questions, skip the redundant one. The aim is a *hazy but
honest* picture in user terms, not a full requirements doc.

A reasonable starting set:

1. **Who is this for?** ("me, when I'm tired on a Friday afternoon" is
   a perfectly valid answer.)
2. **What problem does it solve for them?** — described the way the
   user would describe it, not the way the code implements it.
3. **What would make you scrap this and start again?** — surfaces the
   hidden values that constrain design decisions.
4. **What is it deliberately *not*?** — the negative space matters as
   much as the positive.
5. **If you had to give it a one-line pitch, what would it be?**

Five is plenty. Three can be enough if the user is being terse — don't
drag it out.

## Drafting the entry

Pull the answers together into a short markdown body. Aim for
something a fresh agent could read in under a minute and walk away
with the right mental model. Keep the user's own voice where you can —
direct quotes are good; they age better than your paraphrase.

Show the draft to the user before saving:

> Here's what I'd capture as the foundation. Want me to save this as-is,
> tweak it, or scrap it?

Once they're happy, write the body to a temp file (heredocs are
fragile if the body contains backticks) and run:

```bash
ant add --kind foundation \
        --title "<project> — vision" \
        --body @/tmp/foundation.md
```

Then delete the temp file. The foundation lives in the database now.

## After saving

Mention to the user that the DB is gitignored, so if they want a
durable record of the vision — e.g. committed alongside the code, or
shared in a PR — they can run `ant export <id>` later and save the
markdown wherever feels right.
