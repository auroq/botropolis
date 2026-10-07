# beacon

Shell tooling for the coastal lighthouse network.
Each beacon reports its lamp state, the last time it was heard from and its battery level, and these scripts turn that into something a keeper can act on.

## Scripts

| script | does |
| --- | --- |
| `bin/beacon-status` | one line per beacon: `OK`, `WARN` or `FAIL`, and why |
| `bin/beacon-report` | the daily report: counts by state, who needs a visit, average battery |
| `bin/beacon-rotate` | rotates the lamp logs in `data/logs/` and keeps the last seven |

Every script takes `--help`.

## Quick start

```sh
make status
make report
make rotate     # a dry run; call bin/beacon-rotate directly to rotate for real
make test
```

`BEACON_STATUS_FILE` and `BEACON_LOG_DIR` point the scripts at other data.
`BEACON_NOW` pins the clock, which is how the tests stay deterministic.

## Exit codes

`beacon-status` exits 0 when every beacon is `OK`, 1 when any is `WARN`, and 2 when any is `FAIL`.
All scripts exit 64 on a usage error.

## Tests

`test/run.sh` is plain bash with no framework.
It runs every `test_*` function in its own subshell and scratch directory, and takes an optional pattern to run a subset.
