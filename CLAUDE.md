# Botropolis

Go + Ebitengine isometric city that visualises Claude Code sessions.
Districts are projects, buildings are sessions, the power plant is tokens, radio towers are MCP.

## Where writing goes

This is the rule that was never written down, and the roadmap grew to 2,689 lines because of it.

| file | holds | does not hold |
| --- | --- | --- |
| `ROADMAP.md` | the index: phases, what is open, what is waiting on Aria | any item's detail |
| `roadmap/NNN-*.md` | one numbered item each — the report, the diagnosis, the fix, the frame | rules that outlive the item |
| `roadmap/assets/` | crops and references belonging to one item | frames, which go in `docs/screenshots/` |
| `DESIGN.md` | the index, plus the principles | sections of prose |
| `design/*.md` | what exists — architecture, the city, the pipeline, measured cost | lessons |
| `rules/*.md` | one rule each, paid for by a named mistake | anything not yet generalised |

**A file that is an index stays an index.**
If detail is being added to `ROADMAP.md` or `DESIGN.md`, it belongs in a file underneath them instead.
Both were kept short deliberately so that adding an item or a rule does not grow the thing everyone reads first.

**Items are numbered in one series, 1–68 and counting.**
Numbers are never reused and never renumbered — `roadmap/` filenames, commit messages and code comments all cite them.
There are two bug 18s, left as written, because a dated record that gets tidied stops being a record.

**Item status lives in the item's own file**, on the line under its title: `Done`, `Fixed`, `Open`, `Unadjudicated`, with the date and revision.
`ROADMAP.md`'s open table is derived from those lines, so it is checked rather than remembered — a hand-maintained summary of machine-checkable state always drifts, which it did.

**`make docs-check` enforces all of this**, and runs as part of `make lint`.
It verifies that every relative link resolves *from its own file's directory*, that no `§N` citation to the pre-split roadmap survives, that every item declares a status, and that `ROADMAP.md`'s open table lists exactly the items whose status is `Open`.
It also refuses an empty corpus — the four indexes must be non-empty and every item number from 1 to the highest must have a file — because the other checks all pass on nothing.

Every check was mutated red before the script was trusted, and four of its own versions could not fail:
the first link check tried the bare path before the relative one, so every root-level target resolved from the repo root;
the first corpus check passed on a directory containing four empty files and no `roadmap/`;
and the version after that took its ceiling from the working tree, so deleting the two highest-numbered items lowered `max()` and it reported "1-66, none missing".
The ceiling now comes from `git ls-files`, which a working-tree deletion cannot move — but its `except: pass` silently restored the fault whenever git could not be reached, and the summary then read exactly like a healthy run.
Running outside a checkout is legitimate, so the fallback stays; **it just is not silent any more**, and the green says which source the bound came from.

The pattern across all four is one thing: **a guard fails where its own inputs come from.**
A fallback is a way for a checker to pass, an empty corpus satisfies every check that iterates over it, and a bound derived from the thing it measures can be lowered by the fault it exists to catch.
Mutating a guard tests the guard; mutating what the guard measures itself *against* tests whether it has a fixed reference at all.
A fifth check exists because the other four all pass on nothing — an empty tree reported "0 links resolve, 0 items carry a status, open table matches" and exited 0.
It requires the four index files to be non-empty and every item number from 1 to the highest present to have a file, which is a floor that rises with the work rather than a constant that goes stale.

## Working arrangement

Two sessions, at Aria's request.
One owns `ROADMAP.md`, `roadmap/`, numbered items, audits and relaying to Aria; the other writes code, renders frames to `docs/screenshots/` and packages to `~/workspaces/aur/botropolis-git`.
Aria judges the frames.
Relaying to Aria is the roadmap session's half — check before sending, or she gets the same thing twice.

## Gotchas that have each cost a session

- **`log.showSignature=true` is set repo-locally**, so `git log --oneline | wc -l` counts double and `git log --format=%s` returns the signature ahead of the subject.
  Pass `--no-show-signature` in anything scripted; `git rev-list --count` and `git rev-parse` are unaffected, which is why `pkgver()` has never broken.
- **Repo-local identity and signing**: `Aria Vesta <dev@ariavesta.com>`, SSH signing with `~/.ssh/id_ariavesta_sign_ed25519`.
  Any new repo for this project needs the same four `git config` lines before its first commit, or the global `commit.gpgsign` signs with the work key.
- **The atlas is embedded from `pkg/assets/kits`**, not `pkg/render`, which is the package that draws it.
- **Anything counting sprite usage must exclude `_test.go` and must match string literals rather than file text.**
  A corpus containing the test always agrees with the test, and a doc comment naming a piece is not a caller. See `rules/a-green-test-is-not-a-guard.md`.
- **Restarting the daemon does not restart the map.** When a frame disagrees with the screen, check for a client process first.
