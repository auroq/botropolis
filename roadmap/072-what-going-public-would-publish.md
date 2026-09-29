# 72. Going public would publish the employer's internal repository structure, and history rewriting is the only way back

**Open, and it blocks the publish. Filed 2026-09-29 when Aria chose "let's go public and also use it and let use drive work". Her call, but she should make it knowing this.**

The repo is private today and phase 12's publish step is deliberately blocked on that.
Flipping it is one click and is not reversible in the way it sounds: once public, the clone and the history are out.

## What is in the tree

**`docs/usage-profile.md` is the problem, and it is not about Aria's privacy — it is her employer's.**
Lines 54 onward are a raw per-repository dump from `tools/analyze-history.py`:

```
 78 sess  active   68h21m  out  12.6M  cacheRead  3.7B  PRs 13  last  0d ago  ~/workspaces/github/mCedar/mullet
 10 sess  active   25h22m  out   5.0M  cacheRead  1.6B  PRs 12  last  6d ago  ~/workspaces/github/mCedar/ledger
 27 sess  active   14h08m  out   4.0M  cacheRead  794.3M  PRs 19  last  0d ago  ~/workspaces/github/mCedar/cinders
  4 sess  active   12h30m  out   1.8M  cacheRead  329.1M  PRs  0  last  0d ago  ~/workspaces/bitbucket/tidalUpland/driftingNettle
  2 sess  active   10h40m  out   1.2M  cacheRead  594.2M  PRs 21d ago  ~/workspaces/bitbucket/cinderPylon/emberYard
```

That is **internal repository names, effort distribution across them, and PR throughput per repo**, plus session titles quoting real ticket numbers (`'Ticket 604'`, `'Ticket 602'`) further down.
Nothing in it is dangerous to Aria. All of it is the employer's shape, published under her name, and it is the kind of thing that is nobody's to publish unilaterally.

**Twenty-eight test files use `mullet` as a fixture project name.** Much milder — a string in a fixture, not a dump — but pervasive enough that scrubbing it is a real edit rather than a one-line deletion, and `pkg/commands/status.go:50` gained a doc comment mentioning two mullet sessions *today*, from item 71's investigation. So the tree is still accumulating these.

**The 84 frames in `docs/screenshots/` render district plates, and a district is a project.** Item 40 made the plate permanent by design, so any frame with districts in it shows project names. They are 48 MB and they are the project's whole visual record.

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
