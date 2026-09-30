# 72. Going public would publish the employer's internal repository structure, and history rewriting is the only way back

**Open, and it blocks the publish. Filed 2026-09-29 when Aria chose "let's go public and also use it and let use drive work". Her call, but she should make it knowing this.**

The repo is private today and phase 12's publish step is deliberately blocked on that.
Flipping it is one click and is not reversible in the way it sounds: once public, the clone and the history are out.

## How it got here

`tools/analyze-history.py` scans `~/.claude/projects`, which is **every directory Aria has ever run Claude in** — all of her the employer work included — and prints per-project and per-session tables.
Its output was committed as `docs/usage-profile.md` in this repo's **first commit**, `7a8ecc0` (2026-09-16, *"Design and first-pass plan for Botropolis"*).

It was not a stray paste. The file's fourth line says *"It is what the design in DESIGN.md is shaped around"*, and that is true: the design needed the aggregates — CLI-only, bimodal session length, concurrency, context as the scarce resource — and those are the headline bullets at the top.
**The per-project and per-session tables came along with them**, because the tool prints everything and the whole output was kept as the evidence.

The rest followed from working on the real machine: the Go tests took the paths that were in front of whoever wrote them, and the frames are renders of Aria's actual city.
`sessions.json`, which the same tool writes, was **never committed** — checked across all branches.

## What is in the tree

**1. `docs/usage-profile.md` — the dump, in three regions.**

*Lines 53–73, "Per project":* twenty rows of `~/workspaces/github/mCedar/{mullet, ledger, cinders, terraform, claude-pl…}` and `~/workspaces/bitbucket/{press/main, ops/infrastructure, ops/chef}`, each with session count, active hours, output tokens, cache reads, **PR counts** and recency. This is an inventory of the employer's repositories with effort and throughput against each.

*Line 115, "Skills":* `fallowInlet:hazelInlet` (668 uses), `amberPylon:graniteUpland` (426), `amberPylon:umberEstuary` — internal tooling names with usage counts.

*Lines 129–144, "Top 15 sessions by active time":* session titles, which are ticket identifiers — `'Ticket 604'`, `'Ticket 602'`, `'Issue 597'`, `'Issue 613'`, `'PROJ-1002'`, `'PROJ-1001 Medi…'`, `'Pull request 468…'` — each with hours, prompt counts and tokens. **`PROJ-1001` beside `bitbucket/tidalUpland/driftingNettle` maps an issue-tracker key to its repository.**

**2. Go sources — 71 occurrences of `mullet`/`mCedar` across 29 files, and 292 of `cinders`.**
These are not neutral fixtures: they hardcode real paths such as `/home/avesta/workspaces/github/mCedar/cinders`, and `pkg/city/city_test.go:444` uses PR `#1181` in `mCedar/mullet`.
`pkg/commands/status.go:50` — production, not a test — gained a comment naming two mullet sessions **on 2026-09-29**, out of item 71's investigation, so the tree is still accumulating these.

**3. `docs/screenshots/` — verified by looking, not assumed.**
`r213-project-plates.png` shows district plates reading **`mullet`**, `avesta`, `bot-crossing`, `botropolis`, and the strip across the top reads **`~$192.74 24h`** — a real daily spend — beside `4 need you`, `5 prs`, `3 errors`.
Roughly 45 of the 84 frames are full-window (1100x760, 2200x1520, 1920x1120) and so carry the strip; the rest are crops. Item 40 made the district plate permanent by design, so this is what the program is supposed to draw.

**4. All 323 commits**, since it has been there since the first one. Nothing worse is hiding in deleted history — the deleted paths are Kenney assets and superseded code.

## Why this collides with the item 58 ruling

Aria ruled item 58 **leave it** — no history rewrite.
Item 58's own text says the trigger to revisit is *the repo going public*, and that going public is also **the last moment a rewrite is cheap, before there are clones to invalidate**.

So the two answers meet here: **scrubbing the working tree does not unpublish anything, because all 322 commits go with the repo.** Removing the dump from history means a rewrite, which is exactly the thing she ruled against, at exactly the moment its own filing named as the one to reconsider.

## The options, and none of them is free

1. **Rewrite history now, then publish.** The only option that actually removes it. Cheapest it will ever be — one clone, hers. Invalidates every commit hash, including the ones cited throughout `roadmap/`, `design/` and `rules/`, which is a real cost to a project whose records are half its value.
2. **Publish a fresh repo with no history.** A single squashed initial commit from the current tree, scrubbed. Keeps the private repo intact as the record, publishes a clean one. Loses the commit-level history publicly, which for this project is a genuine loss — the reasoning is in the messages.
3. **Scrub the tree, publish with history, accept what is in it.** Fastest, and the dump stays reachable forever in the history of a public repo. Not recommended, and the reason to state it is that it is what happens by default if nobody decides.
4. **Stay private, publish only the package.** The AUR is blocked on upstream being private, so this means not publishing at all.

**Recommendation: (2).** The history's value is to the two of us and to Aria, not to a reader, and it stays available privately. A squashed public repo costs one afternoon, needs no rewrite of the private record, and the commit hashes cited across `roadmap/` keep resolving in the repo where they were written.

**Whatever she picks, `docs/usage-profile.md` has to be rewritten to aggregates before anything is published** — the headline findings at the top of that file are the part the design was actually shaped around, and they are already aggregate. It is only the per-repo table that has to go.
