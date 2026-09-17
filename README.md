# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

Sessions live in Claude Code's own background daemon; Botropolis starts them, opens them in a terminal, parks them, and wakes them, and never writes to `~/.claude` itself.

See [DESIGN.md](DESIGN.md) for the design and the plan, and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

## Status

Milestone 0.
The three binaries build and answer `version`; nothing else runs yet.

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
