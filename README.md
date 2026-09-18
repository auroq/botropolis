# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

Sessions live in Claude Code's own background daemon; Botropolis starts them, opens them in a terminal, parks them, and wakes them, and never writes to `~/.claude` itself.

See [DESIGN.md](DESIGN.md) for the design and the plan, and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

![The city at r92: a plaza with the plant, hall and library, districts on the ring, storage along the south, the tower ridge and the river](docs/screenshots/r92-beyond-parity.png)

![Thirty seconds of the city](docs/botropolis.gif)

## Start here

```
systemctl --user enable --now botropolisd   # the daemon: inotify on ~/.claude plus hook pushes, a unix socket
botropolis install-hooks                    # botropolis-hook into ~/.claude/settings.json (backup kept)
botropolis doctor                           # the daemon, the hooks, a terminal, the claude CLI, the display
source /usr/share/botropolis/botropolis.bash  # and in ~/.bashrc: a plain `claude` now starts in the background and attaches
botropolis                                  # the city
```

The shell helper is the one line that changes how you work: every session becomes a `claude --bg` job you can close, reopen, park and wake from the map,
and the terminal is only ever a view of it.

## What you see

One district per project, one building per session, and every object stands for one datum you can read by hovering it:
a building's height is its context window used, its beacon is its state (needs-you amber, working blue, waiting-on-its-own-watch teal, unattended violet, parked slate),
a rover works at the door while it is mid-turn, a drone circles the roof for each subagent in flight, a flag per PR (green once merged), smoke per API error.
A vacant plot is a session nothing has been typed into yet;
it is never counted, and `botropolis prune` clears it once it has sat for an hour.
The plant on the plaza is the API — click it for the breakdown by model, project and session over an hour, a day or a week;
the towers on the ridge are MCP servers; the library ranks skills; the city hall carries Claude Code's own rollup.
Avenues carry traffic between projects whose sessions message each other or edit each other's files; parked sessions are containers in the storage yard.
The strip along the top is the city's tallies; the sidebar (`b`) is the same thing as a list.
The timeline (`t`) is the daemon's event log, so it survives the window closing:
when the city opens or comes back into focus, "while you were away" lists what needed you or went wrong since it was last seen.

Keys: drag to pan, wheel to zoom, `r` to turn, `f` or `0` to fit, `tab` for the next session that needs you, `enter` to attach it,
`c` for a new session here, `d d` to demolish, `b` sidebar, `x` breakdown, `t` timeline, `/` search, `s` settings, `n` night/day/live, `[` `]` scrub the clock,
`h` hide the UI, `p` save a frame, `?` all of them.

## Running it

Build the package from `~/workspaces/aur/botropolis-git` (`makepkg -f`, then `pacman -U`), or `make build` for `bin/`.
Then:

```
systemctl --user enable --now botropolisd   # the daemon, socket at $XDG_RUNTIME_DIR/botropolis/botropolis.sock
botropolis install-hooks                    # registers botropolis-hook in ~/.claude/settings.json (backup kept)
botropolis status                           # the table, via the daemon
```

`botropolis city --screenshot city.png` renders one frame and exits;
`--headless` runs it on a virtual display (xvfb-run) so no window opens, `--keys n,n,x` presses keys first, and `--record dir --seconds 24` writes frames for a GIF (`make gif`).
The package also installs `botropolis.desktop`, so the city is in your launcher (`i3-dmenu-desktop`, rofi's drun) under "Botropolis".
`botropolis install-hooks --remove` takes the hook out again.
`botropolis status --direct` skips the daemon.
`botropolis events --since 2h` prints the daemon's event log.
`botropolis` opens the city and `botropolis --tui` the terminal table; `botropolis new <dir> [prompt]`, `attach <id>`, `stop <id>`, `resume <session-id>`, `rm <id>` and `prune` wrap the `claude` CLI;
`packaging/botropolis.bash` makes a plain `claude` in a shell start in the background and attach.

Settings come from flags, then `BOTROPOLIS_HOME`, `BOTROPOLIS_SOCKET`, `BOTROPOLIS_TERMINAL`, `BOTROPOLIS_PARKED_DAYS`, `BOTROPOLIS_CODEX_HOME`,
then `~/.config/botropolis/config.{toml,yaml,json}` (`home`, `socket`, `terminal`, `hook_command`, `parked_days`, `codex_home`, `projection`, `render_scale`, `reduced_motion`, `daily_budget_usd`).
`s` in the city opens a settings panel for the same keys and writes the file back;
`?` lists every key: `x` the power breakdown, `t` the timeline, `/` the search, `b` the sidebar, `r` the heading.

A waybar module, for example:

```json
"custom/botropolis": { "exec": "botropolis bar --watch", "return-type": "json", "on-click": "botropolis" }
```

## Thanks

The city is drawn from [Kenney](https://kenney.nl)'s 3D kits — City Kit Commercial, Roads, Industrial and Suburban, and the Car Kit, all CC0 —
pre-rendered once in Blender by `tools/render-sprites` into the atlases under `pkg/assets/kits/` (four headings, two zoom levels),
the way Factorio ships its sprites: the runtime only ever draws 2D.
A session's building grows with its context window; a rover works at its door while it is mid-turn and a drone circles the roof for each subagent in flight;
a parked session is a shipping container in the storage district;
the plant, city hall and library stand on the plaza; avenues, lamps and trees are placed by the city plan.
`r` turns the camera a quarter at a time.
The 16 px Tiny Town, Tiny Factory and Roguelike Modern City packs remain behind `--projection top`.
Thank you, Kenney; see [`pkg/assets/kenney/README.md`](pkg/assets/kenney/README.md) and [`pkg/assets/kits/README.md`](pkg/assets/kits/README.md) for the packs, versions and terms,
and [support the studio](https://kenney.nl/donate) if the work helps you too.

The chrome is set in [Inter](https://rsms.me/inter/) by Rasmus Andersson, embedded from `pkg/assets/fonts` under the SIL Open Font License 1.1
(the licence text ships beside the face).
It is drawn through Ebitengine's `text/v2` at four sizes and scaled to the display,
or to `render_scale` (`--render_scale 2` for a 2× frame) when set.

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
