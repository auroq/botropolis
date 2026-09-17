# Botropolis

Every Claude Code session on this machine, drawn as a city.
Districts are projects, buildings are sessions, and everything on the map means one thing you can read by hovering it.

Sessions live in Claude Code's own background daemon; Botropolis starts them, opens them in a terminal, parks them, and wakes them, and never writes to `~/.claude` itself.

See [DESIGN.md](DESIGN.md) for the design and the plan, and [docs/usage-profile.md](docs/usage-profile.md) for the numbers it is built around.

## Status

Milestone 1: the read model.
`botropolis status` prints one row per live session straight from `~/.claude`:
state, project, title, branch, model, context used, fresh and cache-read tokens per hour, subagents in flight, and age.
No daemon, no hook, no map yet.

```
$ botropolis status
STATE      PROJECT     TITLE                     BRANCH              MODEL              CTX  FRESH/H  CACHED/H  SUBS  AGE
working    botropolis  scaffold milestone setup  main                claude-opus-5[1m]  29%  152k     8.9M      0/0   4h54m
needs-you  cinders     pr-reviews-cli-migration  feat/cli-pr-review  claude-opus-5[1m]  73%  316k     24.5M     0/4   10h05m
```

`pkg/claude` reads session records, transcripts (metadata, deduplicated usage, cost, whose turn it is), subagents,
`stats-cache.json`, and the MCP config.
`pkg/state` joins them by session id, probes the pid and the background job's pty socket (so an attached `claude --bg` counts as working, not unattended), and derives the state.

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
