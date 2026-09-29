# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

![The city at r136: the plaza with the plant, hall and library, districts on the ring, the storage yard along the south, the tower ridge along the north and the river down the east](docs/screenshots/r136-hero.png)

![Thirty seconds of the city](docs/botropolis.gif)

## Start here

```
systemctl --user enable --now botropolisd   # the daemon: inotify on ~/.claude plus hook pushes, on a unix socket
botropolis install-hooks                    # botropolis-hook into ~/.claude/settings.json (backup kept)
botropolis doctor                           # the daemon, the hooks, a terminal, the claude CLI, the display
botropolis                                  # the city
```

Then one line in `~/.bashrc`:

```
source /usr/share/botropolis/botropolis.bash
```

That line is the one that changes how you work.
Every session becomes a `claude --bg` job you can close, reopen, park and wake from the map, and the terminal is only ever a view of it.

Sessions live in Claude Code's own background daemon;
Botropolis starts them, opens them in a terminal, parks them and wakes them, and never writes to `~/.claude` itself.

## What you see

One district per project, one building per session, and every object stands for one datum.
A building's height is its context window used and its beacon is its state — needs-you amber, working blue, waiting teal, unattended violet, parked slate.
A rover works at the door while a session is mid-turn, a drone circles the roof for each subagent in flight, a flag flies per PR (green once merged) and smoke rises per API error.
A vacant plot is a session nothing has been typed into yet: never counted, and `botropolis prune` clears it after an hour.

The plant on the plaza is the API — click it for the breakdown by model, project and session over an hour, a day or a week.
The towers on the ridge are MCP servers, the library ranks skills, and the city hall carries Claude Code's own rollup.
Parked sessions are containers in the storage yard, one colour per project.
Avenues carry traffic between projects whose sessions message each other or edit each other's files.
The wires from the plant carry each session's live token rate;
a session with no wire is one the daemon has seen no hook events from, so it is reading files.
The freight loop round the city is the ledger: one train per model, a wagon per unit of the day's tokens.
A new session arrives on a tug up the river and its building rises once it docks;
one you demolish leaves downriver.

The strip along the top is the city's tallies and the sidebar (`b`) is the same thing as a list.
The timeline (`t`) is the daemon's event log, so it survives the window closing:
when the city opens or comes back into focus, "while you were away" lists what needed you or went wrong since it was last seen.

Keys: drag to pan, wheel to zoom, `r` to turn, `f` to fit, `tab` for the next session that needs you, `enter` to attach it,
`c` for a new session here, `d d` to demolish, `b` sidebar, `x` breakdown, `t` timeline, `/` search, `s` settings, `n` night/day/live, `[` `]` scrub the clock,
`h` hide the UI, `p` save a frame, `?` all of them.

## Running it

Build the package from `~/workspaces/aur/botropolis-git` (`makepkg -f`, then `pacman -U`), or `make build` for `bin/`.
The package also installs `botropolis.desktop`, so the city is in your launcher under "Botropolis".

`botropolis` opens the city and `botropolis --tui` the terminal table.
`botropolis new <dir> [prompt]`, `attach <id>`, `stop <id>`, `resume <session-id>`, `rm <id>` and `prune` wrap the `claude` CLI;
`status` prints the table, `events --since 2h` the daemon's log, and `install-hooks --remove` takes the hook out again.
`--direct` skips the daemon.

`botropolis city --screenshot city.png` renders one frame and exits;
`--headless` runs it on a virtual display so no window opens, `--keys n,n,x` presses keys first, and `--record dir --seconds 24` writes frames for a GIF (`make gif`).

Settings come from flags, then `BOTROPOLIS_HOME`, `BOTROPOLIS_SOCKET`, `BOTROPOLIS_TERMINAL`, `BOTROPOLIS_PARKED_DAYS`, `BOTROPOLIS_CODEX_HOME`,
then `~/.config/botropolis/config.{toml,yaml,json}` (`home`, `socket`, `terminal`, `hook_command`, `parked_days`, `codex_home`, `projection`, `render_scale`, `reduced_motion`, `daily_budget_usd`).
`s` in the city opens a settings panel for the same keys and writes the file back.

A waybar module, for example:

```json
"custom/botropolis": { "exec": "botropolis bar --watch", "return-type": "json", "on-click": "botropolis" }
```

See [DESIGN.md](DESIGN.md) for the design and the rules, [ROADMAP.md](ROADMAP.md) for where it is going,
and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

## Thanks

The city is drawn from [Kenney](https://kenney.nl)'s 3D kits — City Kit Commercial, Roads, Industrial and Suburban, the Nature Kit, the Car Kit, the Space Kit, the Train Kit and the Watercraft Kit, all CC0 —
pre-rendered once in Blender by `tools/render-sprites` into the atlases under `pkg/assets/kits/` (four headings, two zoom levels),
the way Factorio ships its sprites: the runtime only ever draws 2D.
The 16 px Tiny Town, Tiny Factory and Roguelike Modern City packs remain behind `--projection top`.
Thank you, Kenney; see [`pkg/assets/kenney/README.md`](pkg/assets/kenney/README.md) and [`pkg/assets/kits/README.md`](pkg/assets/kits/README.md) for the packs, versions and terms,
and [support the studio](https://kenney.nl/donate) if the work helps you too.

The chrome is set in [Inter](https://rsms.me/inter/) by Rasmus Andersson, embedded from `pkg/assets/fonts` under the SIL Open Font License 1.1
(the licence text ships beside the face).

## Tools

`tools/analyze-history.py` summarises Claude Code usage from `~/.claude/projects`.
Read-only, stdlib only.

`tools/make-fixtures.py` copies a scrubbed slice of `~/.claude` into `testing/helpers/fixtures/<name>/home/`, laid out like a `$HOME` so tests can point the read model at it.
Ids, timestamps, paths and `usage` are kept;
prompts, tool text, titles and account identifiers become deterministic placeholders.
Fixtures are not committed:
`make test-integration` and `make test-acceptance` regenerate the default `sample` fixture first, and the tests skip on a machine without `~/.claude`.

Both take `--help`.
