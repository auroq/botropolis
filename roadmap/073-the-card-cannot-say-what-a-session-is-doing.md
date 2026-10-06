# 73. The card cannot say what a session has been doing

**Done 2026-10-06, r347, released as 0.1.3. Filed the same day at Aria's request, r344 (`8e6fb88`).**

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

## Ruling 5: the generator is the user's own CLI, 2026-10-06

**`claude -p --no-session-persistence --bare`, and nothing else.**
Aria chose this over the two SDKs after weighing both.

- **The Claude Agent SDK is ruled out.** It exists only in Python and TypeScript, and it runs the Claude Code binary as a subprocess.
  It shows up on the machine exactly as the CLI does, and it would bring a second runtime into a Go app.
- **`anthropic-sdk-go` is ruled out.** It would be invisible to the city, but it needs an API key.
  Reusing the Claude Code login from inside botropolis is not an option: the Agent SDK overview says *"Unless previously approved, Anthropic does not allow third party developers to offer claude.ai login or rate limits for their products"*.
  The repo is headed for public release and the AUR, so reading `~/.claude/.credentials.json` would be exactly that.
- **The CLI uses the user's existing login without botropolis ever touching it.** The user's own installed Claude Code makes the call, which is what the usage probe already does.

**`--bare`'s scope is disputed.**
The docs at `code.claude.com/docs/en/cli-reference.md` describe it as skipping hook *auto-discovery*.
The installed `claude --help` says it skips hooks "defined in settings and by installed plugins", which is stronger.
The session-count test above settles that, so neither source has to be trusted.

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

## What building it found, 2026-10-06

**Ruling 5's `--bare` was wrong, and the ruling's own reason is what rules it out.**
The full `claude --help` text says that under `--bare` *"Anthropic auth is strictly ANTHROPIC_API_KEY or apiKeyHelper via --settings (OAuth and keychain are never read)"*.
The help I quoted when the ruling was made was cut short before that sentence.
So `--bare` needs the API key that using the CLI was chosen to avoid.
The ruling's substance stands — the user's own CLI and login, no SDK — and the flags that deliver it are:

| flag or fence | what it closes |
| --- | --- |
| `--no-session-persistence` | the transcript, so no parked session afterwards |
| running in `SummaryDir()`, fenced in `state.Build` beside the usage probe | the live session record a print run writes to `~/.claude/sessions/<pid>.json` while it runs — the third path, which the table above left unverified, is real |
| `--setting-sources ""` | the user's hooks and plugins firing for a run they did not start |
| `--tools ""` | anything but an answer |
| `--strict-mcp-config` | see below |

Hooks were never a way onto the map: `Daemon.Apply` only overlays sessions the loader already has, so a hook event for an unknown id creates nothing.

**The acceptance test was run, and it is what proved the fence.**
A real summary was generated while two builds counted sessions from the same `~/.claude`: this change held at 31 before, during and after, and a build of `HEAD` without the summary fence read **32 for the whole run**.

**Without `--strict-mcp-config` it failed outright.**
Every connected MCP server's tool definitions ride along, and on Aria's machine that was ~320k tokens of prompt — over the window — for a one-paragraph answer.
With it, ~6.5k.
The first live run happened to succeed, presumably before the connectors had loaded, which is how a run that fails every time looked like it worked once.
Claude Code reports that failure on **stdout**, not stderr, so the error now reads both.

**An unpersisted run still leaves an empty project folder** with an empty `memory/` in it.
`Summarise` removes both afterwards with `os.Remove`, which refuses a non-empty directory, so it cannot take anything that is not empty.
The usage probe's folder carries the same empty `memory/`; left alone, since it also holds the one transcript item 67 keeps on purpose.

**The snapshot note above was wrong in a useful direction.**
`harness.Multi.Load` copies `Session` values whole; the trap is for new `Snapshot` fields.
`Recap` and `LastPrompt` are on `Session` and reach the client without a merge change.

**As built:**

- `claude.Recap` (text, time, prompts since) and `LastPrompt` are read from the transcript, and carried on `state.Session`; `status --json` shows them.
- The card's **summary** button, or `i`, toggles a wrapped paragraph — `ui.ParaWidth`, 60 grid units — under an age line.
  A new selection always starts closed.
- A generated summary is held in the map process only, and is gone on restart.
  It is labelled as generated, with its age and the model.
- `Generate` runs off the frame, one per session at a time, and hands its result back to `Update`.
- Codex sessions get no summary button: they have neither a recap nor a transcript the excerpt reader understands.
- The excerpt sent is the main line's typed prompts and replies from the transcript's last 2 MB, newest 24 KB of them, on stdin rather than the command line.

**Frames:** rendered headlessly from a copy of the scrubbed fixture with a hand-written recap injected, its session records pointed at a live PID so `tab` had something to select.
Not committed: the tree keeps one frame on `make publishable`'s reviewed list, and these were sent to Aria directly.

**Not done:** the TUI has its own card and was not given a summary.

