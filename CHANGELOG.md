# Changelog

All notable changes to Botropolis are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-09-30

### Fixed

- The `.deb` and `.rpm` packages no longer require `libXxf86vm` or ALSA.
  Neither is referenced by any binary — they came from the build headers CI installs, which is not the same as a runtime need, and requiring ALSA made every user install a sound stack for a program that never opens one.
- The packages now declare `hicolor-icon-theme`, which they need for the icon they install into that tree.

## [0.1.0] - 2026-09-30

First release.
Botropolis draws every Claude Code session on the machine as a city, and starts, opens, parks and wakes them through the `claude` CLI.

### Added

- `botropolis` — the isometric city.
  One district per project, one building per session, and every object stands for one datum you can read by hovering it.
  A building's height is its context used and its beacon is its state;
  rovers, drones, flags and smoke stand for turns in flight, subagents, pull requests and API errors.
- The plaza landmarks — the power plant for API spend, broken down by model, project and session over an hour, a day or a week;
  radio towers for MCP servers;
  the library for skills;
  and the city hall for Claude Code's own rollup.
- The river, the freight loop and the avenues — arrivals and departures, the day's tokens as one train per model, and traffic between projects whose sessions message each other or edit each other's files.
- Nine info views, a sidebar, a breakdown panel, a searchable list and a timeline that survives the window closing,
  including a "while you were away" summary of what needed you since the city was last seen.
- `botropolisd` — a daemon watching `~/.claude` by inotify, taking hook pushes, and serving the city model over a unix socket.
- `botropolis-hook` — the hook forwarder, installed into `~/.claude/settings.json` by `botropolis install-hooks` with a backup kept.
- Session management wrapping the `claude` CLI: `new`, `attach`, `stop`, `resume`, `rm` and `prune`.
  Botropolis never writes to `~/.claude` itself.
- `botropolis status`, `bar` for a status bar such as waybar, `tui` for the terminal table, `events` for the daemon log, and `doctor` for the install.
- `botropolis version`, and `--version` on all three binaries.
- Settings from flags, environment and `~/.config/botropolis/config.{toml,yaml,json}`, with an in-city settings panel that writes the file back.
- Headless rendering — `--screenshot` for one frame, `--record` for a sequence, and `--keys` to press keys first.
- `.deb`, `.rpm` and Arch packages, each installing the three binaries, both systemd user units, the desktop entry, the icon, the shell integration and the licences.

### Security

- The usage probe runs inside a fenced directory the session loader skips, and prunes its own transcripts, so reading usage cannot make Botropolis see itself as a session.

[Unreleased]: https://github.com/auroq/botropolis/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/auroq/botropolis/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/auroq/botropolis/releases/tag/v0.1.0
