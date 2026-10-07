# Demo media

Screenshots and videos of the city are filmed from demo sessions, never from a real `~/.claude`.
Twice before, frames of a real city published things they should not have: item 72 is the record.
This directory is the way around that.

## How it fits together

| path | holds |
| --- | --- |
| `corpus/` | recorded transcripts, one folder per toy project, as Claude Code wrote them |
| `scenarios/` | one city each: which sessions stand where, in what state, at what time of day, and the shots to make of it |
| `Dockerfile` | the sealed containers the demo is filmed and recorded in |

`botropolis-demo stage` builds a disposable home from the corpus for one scenario.
It moves every transcript so it ends as long ago as the scenario says,
because the power plant, the trains and the storage yard only see the last day or week.
It writes a session record for each live session, pointed at a detached `sleep`,
because the city decides a session is live by whether its pid answers.

`botropolis-demo film` stages each scenario afresh for every shot, runs the city headless against it, and encodes clips.
It writes `dist/demo/manifest.json` listing every file with its title, description, size and duration.

## Filming

```sh
make demo-media
```

That builds the binaries and the renderer image, then films every scenario inside a container.
The container has no network and a read-only root, and it mounts only the binaries, the corpus, the scenarios and the output directory.
`HOME`, the XDG directories and `TZ` all point inside it.
Nothing the city does can reach the real home, including the reads that would otherwise fall back to it.

`DEMO_CORPUS`, `DEMO_SCENARIOS` and `DEMO_OUT` point it somewhere else.

## A scenario

```yaml
name: hamlet
description: One person, two projects, late evening.
clock: "22:30"          # the light follows this; 14:00 when unset
mcp: [tracker, weather] # radio towers
sessions:
  - ref: tidepool/6f1d  # <project>/<session id or a prefix of it>
    state: working      # working, needs-you, unattended, parked
    ago: 2m             # how long before now its last line sits
  - project: beacon     # an empty plot needs no recording
    state: empty
shots:
  - name: overview
    title: A quiet hamlet
    description: Two projects, one session working.
  - name: tour
    keys: [equal, equal, tab]
    record: {seconds: 12, fps: 30}
    formats: [mp4, webm, gif]
```

A shot also takes `window`, `hover`, `projection`, `detail`, `signage`, `scale` and its own `clock`.
Unknown keys are refused, and so is a scenario that places no sessions.
