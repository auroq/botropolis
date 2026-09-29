# 71. Two live sessions are missing from status entirely, and one was active today

**Closed 2026-09-29 as a false positive, same day it was filed. Both sessions were present and always were.** Driving `state.Loader` against the real `~/.claude` finds `d7314bb1` as **"What time is it"** and `ac752824` as **"Database credential setup migration to Chef"**, both `needs-you`, both alive, 20 sessions in the snapshot and 0 files skipped. Verified independently through the `status --json` that r322 added.

`claude agents --json --all` reports 13 sessions as `blocked` and 2 as `working`.
`botropolis status` shows 12 rows, `--all` shows 20 (9 needs-you, 3 working, 8 parked).
Reconciling the two, **eight of the CLI's live sessions are not in `status`**, and six of those eight are explained:

- three are parked and appear under `--all` — correct;
- three are idle 7.9, 8.8 and 10.7 days against `parked_days` of 7, so they are not catalogued at all — also correct.

**Two are not explained.**

| CLI name | sessionId | cwd | CLI state | idle | transcript |
| --- | --- | --- | --- | --- | --- |
| `d7314bb1` | `d7314bb1` | mullet | `blocked` | 0.3 d | 385 KB, 38 lines |
| `7d8de921` | `ac752824` | chef | `blocked` | **0.0 d** | 311 KB, 150 lines |

Neither appears in `status`, in `status --all`, or in `botropolis bar`'s tooltip.
One of them was active on the day it was found.

**What has been ruled out, each by checking rather than by reasoning:**

- **Not `kind`.** Both are `background`, the same as every other live session.
- **Not bridge-session stubs.** 385 KB and 311 KB of real conversation, 38 and 150 records. The *smallest* transcript of the three untitled sessions — `f26cfe27` at 2,445 bytes and 6 records — is the one that **is** shown.
- **Not the untitled name.** `f26cfe27` also has a bare 8-hex id for a title and appears normally.
- **Not `parked_days`.** Both are inside it by a wide margin.
- **Not the project root.** `f26cfe27` and `ac752824` are both under the same `bitbucket` tree, and only one of them shows.

**The one asymmetry noticed and not chased:** `7d8de921`'s CLI *name* is not its own sessionId — the name is `7d8de921` and the session is `ac752824` — where the other two untitled sessions have name equal to sessionId. Whether the loader keys on one and the CLI on the other is worth looking at first, and `d7314bb1` does **not** have that asymmetry, so it cannot be the whole cause.

Also noticed: `d7314bb1`'s transcript opens with an `ai-title` record where the others open with `mode`, which may matter to however the loader decides a session is live.

**This is the build session's, because it is in the loader and I would be guessing at code I do not own.** What it wants is the derivation in `pkg/claude` for which sessions reach a snapshot, checked against these two sessionIds directly rather than against the aggregate.

**And it is the argument for the checklist.** Ten steps, never run end to end, and the first one found a live session missing from the program's main view.


---

## Why the filing was wrong, which is the only part worth keeping

**`claude agents` names a session by its jobId; botropolis titles it from the transcript's `ai-title`.**
The only column the two lists share is the title, so that is what I reconciled on — and titles are not a key.
`status --all` prints **"What time is it" twice in mullet**, two different sessions, and "chef-tools homebrew tap" twice as well.

The two rows I reported as missing sessions were sitting in the output I was looking at, under names the CLI does not use.

**And I had already written the reason down before filing anyway.** The same step-1 notes say *"titles are not unique (two sessionIds share 'chef mac documentation' with different states)"* — I identified the instrument fault, recorded it as a caveat on the comparison, and then filed a defect from that comparison regardless.
That is *name the set before quoting the number*, arriving as a bug report: the set was "rows I could match by title" and it was quoted as "sessions that exist".

Neither asymmetry I flagged was a cause. `ac752824`'s name-is-not-its-sessionId is real and incidental; `d7314bb1`'s `ai-title`-first record is a coincidence.

**What the false positive bought, which is why it is filed rather than deleted:** `status --json` now emits `state.Session` as modelled, carrying `id` and `pid`, so step 1 is a check rather than a judgement and cannot be reconciled on a non-key again. Run by id against the real tree, step 1 passes — see [validation-checklist.md](validation-checklist.md).
