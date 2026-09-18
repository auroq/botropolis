# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

Sessions live in Claude Code's own background daemon; Botropolis starts them, opens them in a terminal, parks them, and wakes them, and never writes to `~/.claude` itself.

See [DESIGN.md](DESIGN.md) for the design and the plan, and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

## Status

Milestone 1: the read model.
`botropolis status` prints one row per live session straight from `~/.claude`:
state, project, title, branch, model, context used, fresh and cache-read tokens per hour, subagents in flight, and age.
Milestone 6: satellites.
`botropolis bar` prints one line for waybar (JSON) or any text bar (`--format text`, `--watch` to stream);
`botropolis notify` (unit `botropolis-notify.service`) sends a desktop notification when a session starts needing you;
`botropolis --tui` (or `botropolis tui`) is the live table in your terminal with attach, stop, resume and demolish keys.
Harnesses sit behind `harness.Snapshotter`: Claude Code is the first, and a Codex CLI adapter reads `~/.codex/sessions` rollouts
(fixture-tested only — there is no Codex on this machine yet).

Milestone 5: parked sessions are catalogued from transcripts with no live record (`parked_days`, default 7) and drawn boarded up;
the power plant sums the last 24 h of `cost-state` per model and runs lines to every lit building (fresh warm and thick, cached cool and thin);
one radio tower per MCP server (configured or merely used) with beams to the sessions that called it;
the library ranks skills; the city hall carries Claude Code's own `stats-cache.json` rollup (lifetime totals, busiest hour, model mix);
roads run between districts whose sessions message each other or edit each other's files, signed with the traffic;
buildings fly a flag per PR and smoke per API error; it is night while anything runs unattended.
Hover a building for what its worker is doing right now (tool and file), context with compaction history, the subagents in flight by name, PRs, errors and the last hook note.
`d d` on a selected building demolishes it (`claude rm`); `botropolis prune [--older-than 168h] [--dry-run]` removes parked background jobs.
`botropolis status --all` lists parked sessions in the table.

Milestone 4: `botropolis` with no arguments (or `botropolis city`) opens the city:
one district per project, one building per session, fill level for context used,
orange and pulsing for needs-you, cranes for subagents in flight, boarded up for parked.
Drag to pan, wheel to zoom, click a building to attach it, `tab` (or a click on a state chip in the strip) to jump to the next session in that state and `enter` to attach it, `f` to fit, `n` to force night (to see the lights), `q` to quit.
The layout is sticky, in `~/.local/state/botropolis/layout.json`.
The window is built for tiling: it fits the city to whatever size the window manager gives it, refits on resize until you pan or zoom (`f` refits),
and wraps the footer and clamps the hover card at narrow or short sizes.
It sets `WM_CLASS` to `botropolis`, so if you would rather float it under i3:

```
for_window [class="Botropolis"] floating enable, resize set 1100 760
```

Milestone 3: `botropolis new/attach/stop/resume/rm` wrap the `claude` CLI.

Milestone 2: `botropolisd` watches `~/.claude` with inotify and serves snapshots over a unix socket;
`botropolis-hook` forwards Claude Code hook events to it in single-digit milliseconds;
`status` asks the daemon first and scans directly when it is down.
No map yet.

```
$ botropolis status
STATE      PROJECT     TITLE                     BRANCH              MODEL              CTX  FRESH/H  CACHED/H  SUBS  AGE
working    botropolis  scaffold milestone setup  main                claude-opus-5[1m]  29%  152k     8.9M      0/0   4h54m
needs-you  cinders     pr-reviews-cli-migration  feat/cli-pr-review  claude-opus-5[1m]  73%  316k     24.5M     0/4   10h05m
```

`pkg/claude` reads session records, transcripts (metadata, deduplicated usage, cost, whose turn it is), subagents,
`stats-cache.json`, and the MCP config.
`pkg/state` joins them by session id, probes the pid and the background job's pty socket (so an attached `claude --bg` counts as working, not unattended), and derives the state.

## Running it

Build the package from `~/workspaces/aur/botropolis-git` (`makepkg -f`, then `pacman -U`), or `make build` for `bin/`.
Then:

```
systemctl --user enable --now botropolisd   # the daemon, socket at $XDG_RUNTIME_DIR/botropolis/botropolis.sock
botropolis install-hooks                    # registers botropolis-hook in ~/.claude/settings.json (backup kept)
botropolis status                           # the table, via the daemon
```

`botropolis install-hooks --remove` takes the hook out again.
`botropolis status --direct` skips the daemon.
`botropolis` opens the city and `botropolis --tui` the terminal table; `botropolis new <dir> [prompt]`, `attach <id>`, `stop <id>`, `resume <session-id>`, `rm <id>` and `prune` wrap the `claude` CLI;
`packaging/botropolis.bash` makes a plain `claude` in a shell start in the background and attach.

Settings come from flags, then `BOTROPOLIS_HOME`, `BOTROPOLIS_SOCKET`, `BOTROPOLIS_TERMINAL`, `BOTROPOLIS_PARKED_DAYS`, `BOTROPOLIS_CODEX_HOME`,
then `~/.config/botropolis/config.{toml,yaml,json}` (`home`, `socket`, `terminal`, `hook_command`, `parked_days`, `codex_home`).

A waybar module, for example:

```json
"custom/botropolis": { "exec": "botropolis bar --watch", "return-type": "json", "on-click": "botropolis" }
```

## Thanks

The city is drawn with tiles by [Kenney](https://kenney.nl) — the isometric Buildings, City, Landscape and Vehicles packs,
Roads, and Tiny Town, Tiny Factory and Roguelike Modern City for the top-down view (`--projection top`), all CC0.
A resource strip along the top keeps the city-wide tallies in view: sessions by state, token rates, cost and cache hit rate over 24 h, MCP calls, PRs, errors.
Streets between districts are autotiled from the road pack with cars for traffic (one per street, up to three with traffic; never faster); power lines run on poles with sparks for token flow; context fill is the number of storeys.
At night — whenever something runs unattended — the map dims and every building with a session awake in it shows lit windows.
A river and a pond sit in the outskirts as background; they mean nothing.
Thank you, Kenney; see [`pkg/assets/kenney/README.md`](pkg/assets/kenney/README.md) for the packs, versions and terms,
and [support the studio](https://kenney.nl/donate) if the work helps you too.

## Tools

`tools/analyze-history.py` summarises Claude Code usage from `~/.claude/projects`.
Read-only, stdlib only; `--help` for options.

`tools/make-fixtures.py` copies a scrubbed slice of `~/.claude` into `testing/helpers/fixtures/<name>/home/`,
laid out like a `$HOME` so tests can point the read model at it.
Ids, timestamps, paths, and `usage` are kept;
prompts, tool text, titles, and account identifiers become deterministic placeholders.
Fixtures are not committed:
`make test-integration` and `make test-acceptance` regenerate the default `sample` fixture first,
and the tests skip on a machine without `~/.claude`.
`--help` for options.
