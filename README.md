# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

Sessions live in Claude Code's own background daemon; Botropolis starts them, opens them in a terminal, parks them, and wakes them, and never writes to `~/.claude` itself.

See [DESIGN.md](DESIGN.md) for the design and the plan, and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

## Status

Milestone 0.
Nothing runs yet.

## Tools

`tools/analyze-history.py` summarises Claude Code usage from `~/.claude/projects`.
Read-only, stdlib only; `--help` for options.
