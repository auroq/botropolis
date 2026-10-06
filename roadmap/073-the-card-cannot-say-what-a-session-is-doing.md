# 73. The card cannot say what a session has been doing

**Open. Filed 2026-10-06 at Aria's request, r344 (`8e6fb88`). A feature, specified with Aria; the build session's to build.**

Clicking a building raises a card with a title, numbers and a row of actions.
Every one of those actions is a way *into* the session.
None of them answers the question that usually comes before going in: what was this thread about, and where did it get to?

**What it wants:** a **Summary** button on the building's card that expands the card to show a summary of the thread.

## The summary already exists, on disk

Claude Code writes the recap that appears at the bottom of a session left waiting for input into the transcript as a record:

```json
{"type": "system", "subtype": "away_summary", "content": "We're finishing … Next: …", "timestamp": "2026-10-06T17:14:37.101Z"}
```

Measured 2026-10-06 over the 60 most recently modified transcripts under `~/.claude/projects`:

- **38 of 60 carry at least one.**
- They accumulate rather than replace: one long session carried 44.
  **The last one is the summary**; the earlier ones are history.
- 350 `away_summary` records across those 60 files, against 4,819 `ai-title` records.

`pkg/claude/transcript.go` already switches on record type for `custom-title` and `ai-title`, so reading the newest `away_summary` is the same shape as reading the title.
Nothing in the repo reads it today.

## Aria's rulings, 2026-10-06

1. **Where it shows: the card expands.**
   The button toggles a word-wrapped paragraph inside the existing card, not a separate panel.
   `LayoutCard` clips every line to the card's width today, so a paragraph needs real wrapping, not more clipped lines.
2. **Staleness: show its age.**
   The recap is shown as written, under a line saying when it was written and how far the session has moved since — "recap from 3h ago, 12 turns since".
   A recap written before the latest prompt is still useful; one that pretends to be current is not.
3. **No recap: fall back, and offer to make one.**
   About a third of sessions have none — mostly short ones, or ones nobody stepped away from.
   For those the expanded card shows the title and the last prompt (`last-prompt` is already in the transcript, 4,833 records in the same sample), labelled as *no recap yet*, plus a **Generate summary** button.
   That way, spending tokens is a choice the user makes when they press it, never something the map does on its own.
4. **Generating one must not put a new building on the map.**

## Ruling 4 is the hard part, and this repo has been here before

The usage probe already shells out to `claude -p`, and it reached the city twice.
[Bug 54](054-the-probe-fence-covers-parked-transcripts.md) fenced its transcripts out of the loader, and [item 67](067-the-usage-probe-s-transcripts-are-never.md) found them piling up and added the prune.
A generator that copies the probe's `exec.CommandContext(ctx, bin, "-p", …)` gets both problems back.

A generated summary has **three** ways to become a session, and each needs closing on its own:

| path | what closes it |
| --- | --- |
| a transcript under `~/.claude/projects` | `--no-session-persistence`, which writes nothing to disk and is better than the probe's fence-and-prune |
| hook events reaching the daemon | `--bare`, which skips settings and plugin hooks; without it the daemon sees a live, hooked session for as long as the run takes |
| whatever else lists live `claude` processes — `~/.claude/sessions/<pid>.json`, `claude agents` | **not verified**; check whether either sees a `-p --no-session-persistence` run before trusting the first two |

**The acceptance test is the city, not the flags.**
Generate a summary while the daemon is running, and assert that the snapshot's session count is the same before, during and after the run.
That catches all three paths and any fourth one nobody has thought of yet.

A generated summary is not an `away_summary` and must not be written into the session's transcript — that file belongs to Claude Code.
Where it is kept, and whether it survives a restart, is the build session's call; it should carry its own timestamp and be labelled as generated, so the age line in ruling 2 still holds.

## Things to settle while building

- **What gets sent.** A whole transcript can be megabytes.
  The prompt wants the tail plus the title, or the `compact_boundary` summary when there is one, rather than the file.
- **Which model.** A summary does not need the session's own model; a small one is cheaper and faster.
- **The wait.** A `-p` run takes seconds.
  The button needs a pending state, and the card must not block the frame loop while it waits — the [10% CPU bar](069-the-10-percent-cpu-bar-is-missed.md) is already missed.
- **The snapshot.** A new `Session` field for the recap must be merged in `harness.Multi.Load`, or the daemon drops it on the way to the client.
  That has cost a session before.
- **Privacy.** The recap is the session's own words, and it can name employer projects — the example above does.
  It goes nowhere new on the user's own screen, but it **must not** appear in any frame committed to `docs/screenshots/` without going through `make publishable`'s reviewed list, which is the trap [item 72](072-what-going-public-would-publish.md) is about.
