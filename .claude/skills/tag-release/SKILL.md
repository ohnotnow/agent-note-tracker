---
name: tag-release
description: >
  Pre-flight check before cutting an ait release. Reports the state of the working tree, the
  latest tag, and CHANGELOG.md, then flags anything that looks off. Does NOT run git, gh, or any
  release commands — just reports findings and hands back to the maintainer. Use when the user
  says "tag a release", "ship a release", "pre-flight check", "tag-release", or asks to look over
  the repo before they cut a tag.
allowed-tools: "Read,Bash"
version: "0.2.0"
---

# tag-release (pre-flight check)

A read-only sanity check before the maintainer cuts a release. Gather facts, flag anything that
looks wrong, hand back. **Do not run `git commit`, `git tag`, `git push`, `gh release ...`, or
edit any files.** The maintainer drives the actual release; this skill just spots problems.

## What to check

Run these (read-only) and gather the output:

- `git status --short` — anything uncommitted or untracked.
- `git branch --show-current` — current branch.
- `git tag --sort=-v:refname | head -3` — most recent tags.
- `git log <latest-tag>..HEAD --oneline` — commits since the latest tag.
- `git show <latest-tag> --stat --format='%s%n%n%b' --no-patch` — the latest tag's annotation and what was in it.
- Top of `CHANGELOG.md` — the `[Unreleased]` block and the latest `## [X.Y.Z]` heading.

If `git status --short` shows untracked directories, peek inside (`ls` or `git ls-files --others --exclude-standard <path>`) so you can describe what's there rather than just "something untracked".

## What to report

A short markdown report with these sections, in this order. Skip a section if there's nothing to
say in it.

### State
A few bullet points: branch, latest tag, commits since latest tag (count is fine, list them if
≤5), `[Unreleased]` content (paste it verbatim if short, summarise if long).

### Flags
Anything the maintainer should notice before tagging. One line each. Examples:

- Latest tag `vX.Y.Z` has no entry in CHANGELOG.md (mismatch).
- `[Unreleased]` is empty but there are uncommitted changes in `<path>` — those might be the release.
- `[Unreleased]` is empty and the tree is clean — nothing to release right now.
- Branch is `<x>`, not `master`.
- Latest tag annotation is empty / says "test" / is a single character.
- GitHub release for `vX.Y.Z` likely shows only the auto-generated "Full Changelog" line because the changelog wasn't updated before tagging.

Don't pad. If everything's fine, say so in one line.

### Suggested next steps
Plain English, not commands. Examples:

- "Add a `[1.8.1]` entry to CHANGELOG.md describing the test tag, then `gh release edit v1.8.1` to match."
- "Commit `.claude/skills/tag-release/` and bump CHANGELOG.md before tagging."
- "Nothing to do — ready to tag whenever you are."

The maintainer knows git and `gh`. **Do not write out command sequences, heredocs, or temp-file
recipes.** A one-line `gh release edit ...` example is fine if it genuinely clarifies; anything
longer belongs in the maintainer's hands, not the skill's output.

## Notes

- Strictly read-only. No edits to CHANGELOG.md, no `gh` writes, no `git` writes.
- Voice: matter-of-fact, brief. The maintainer is sleepy and busy; respect their time.
- If the maintainer asks you to *do* something after the report (e.g. "go ahead and update the changelog"), that's fine — but only after they've explicitly asked.
