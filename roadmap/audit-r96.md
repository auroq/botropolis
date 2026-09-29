## 8. Audit, 2026-09-18 (r96, `138c8ba`)

Phases 7–12 shipped in one day.
The audit built HEAD, ran every command against this machine, shot every panel headlessly, measured the daemon and the hook,
and recomputed the numbers from the raw transcripts.

**Holds up:** `go build`, `go vet`, `golangci-lint` (0 issues), 25 packages of tests green, race detector clean on daemon/state/claude/city;
14.4k lines of Go with 12.2k of tests; coverage 85–97% on every model package (`plan` 97.5, `ui` 94.5, `city` 92.3, `state` 90.6, `claude` 88.9).
Daemon 17.4 MB RSS, 15 threads, ~0.4% CPU idle; hook 3 ms round trip with a daemon, 2 ms and exit 0 without.
`status` context for this session matches the transcript exactly (519,298 tokens, 52% of 1M); parked count 70 vs 71 on disk (the stub threshold).
`doctor` found the terminal gap on its first run and terminator is now a known terminal.
Rotation, zoom, sidebar, breakdown, timeline, settings, F1 help, hide-chrome all render headlessly.

### Bugs, most important first
