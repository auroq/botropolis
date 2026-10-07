#!/usr/bin/env bash
set -euo pipefail

usage() {
    cat <<USAGE
Usage: test/run.sh [--help] [PATTERN]

Run every test_* function in this file whose name matches PATTERN
(default: all). Each test runs in its own subshell and scratch directory.
USAGE
}

[[ "${1:-}" == "-h" || "${1:-}" == "--help" ]] && { usage; exit 0; }
pattern="${1:-}"

HERE="$(cd -- "$(dirname "$0")" >/dev/null && pwd)"
BIN="${HERE}/../bin"
FIXTURES="${HERE}/fixtures"
export BEACON_NOW="2026-10-07T06:00:00Z"

fail() { echo "    $*" >&2; return 1; }

assert_eq() {
    [[ "$1" == "$2" ]] || fail "expected [$2], got [$1]"
}

assert_contains() {
    [[ "$1" == *"$2"* ]] || fail "expected output to contain [$2], got:"$'\n'"$1"
}

assert_file() {
    [[ -e "$1" ]] || fail "expected $1 to exist"
}

assert_no_file() {
    [[ ! -e "$1" ]] || fail "expected $1 not to exist"
}

status_exit() {
    local code=0
    "${BIN}/beacon-status" "$@" >/dev/null 2>&1 || code=$?
    echo "${code}"
}

test_status_when_every_beacon_is_healthy_it_should_exit_0() {
    assert_eq "$(status_exit --file "${FIXTURES}/all-ok.txt")" 0
}

test_status_when_a_lamp_is_not_ok_it_should_exit_2() {
    assert_eq "$(status_exit --file "${FIXTURES}/status.txt")" 2
}

test_status_when_a_lamp_is_dim_it_should_mark_it_failed() {
    assert_contains "$("${BIN}/beacon-status" --file "${FIXTURES}/status.txt" || true)" "FAIL  BX03"
}

test_status_when_a_beacon_is_silent_it_should_warn_with_the_age() {
    assert_contains "$("${BIN}/beacon-status" --file "${FIXTURES}/status.txt" || true)" "silent 120m"
}

test_status_when_the_battery_is_low_it_should_warn() {
    assert_contains "$("${BIN}/beacon-status" --file "${FIXTURES}/status.txt" || true)" "battery 12%"
}

test_status_when_the_stale_window_is_widened_it_should_not_warn() {
    local out
    out="$("${BIN}/beacon-status" --file "${FIXTURES}/status.txt" --stale-minutes 180 || true)"
    [[ "${out}" != *"silent"* ]] || fail "still reported silent beacons"
}

test_status_when_quiet_it_should_omit_ok_beacons() {
    local out
    out="$("${BIN}/beacon-status" --file "${FIXTURES}/status.txt" --quiet || true)"
    [[ "${out}" != *"BX01"* ]] || fail "printed an OK beacon"
}

test_status_when_the_file_is_missing_it_should_exit_64() {
    assert_eq "$(status_exit --file "${SCRATCH}/nope.txt")" 64
}

test_report_when_run_it_should_count_beacons_by_state() {
    assert_contains "$("${BIN}/beacon-report" --file "${FIXTURES}/status.txt")" "beacons: 5   ok: 3   warn: 1   fail: 1"
}

test_report_when_run_it_should_average_the_battery() {
    assert_contains "$("${BIN}/beacon-report" --file "${FIXTURES}/status.txt")" "average battery: 52%"
}

test_report_when_a_date_is_given_it_should_put_it_in_the_header() {
    assert_contains "$("${BIN}/beacon-report" --file "${FIXTURES}/status.txt" --date 2026-01-31)" "report for 2026-01-31"
}

test_report_when_nothing_needs_a_visit_it_should_say_so() {
    assert_contains "$("${BIN}/beacon-report" --file "${FIXTURES}/all-ok.txt")" "No visits needed."
}

make_logs() {
    mkdir -p "${SCRATCH}/logs"
    echo "today" > "${SCRATCH}/logs/carrick.log"
    echo "yesterday" > "${SCRATCH}/logs/carrick.log.1"
    echo "oldest" > "${SCRATCH}/logs/carrick.log.2"
}

test_rotate_when_run_it_should_shift_the_current_log_to_1() {
    make_logs
    "${BIN}/beacon-rotate" --dir "${SCRATCH}/logs" --keep 3 >/dev/null
    assert_eq "$(cat "${SCRATCH}/logs/carrick.log.1")" "today"
}

test_rotate_when_run_it_should_leave_an_empty_current_log() {
    make_logs
    "${BIN}/beacon-rotate" --dir "${SCRATCH}/logs" --keep 3 >/dev/null
    [[ ! -s "${SCRATCH}/logs/carrick.log" ]] || fail "current log is not empty"
}

test_rotate_when_past_keep_it_should_drop_the_oldest() {
    make_logs
    "${BIN}/beacon-rotate" --dir "${SCRATCH}/logs" --keep 2 >/dev/null
    assert_eq "$(cat "${SCRATCH}/logs/carrick.log.2")" "yesterday"
}

test_rotate_when_dry_run_it_should_change_nothing() {
    make_logs
    "${BIN}/beacon-rotate" --dir "${SCRATCH}/logs" --dry-run >/dev/null
    assert_eq "$(cat "${SCRATCH}/logs/carrick.log")" "today"
}

test_rotate_when_keep_is_zero_it_should_refuse() {
    make_logs
    local code=0
    "${BIN}/beacon-rotate" --dir "${SCRATCH}/logs" --keep 0 >/dev/null 2>&1 || code=$?
    assert_eq "${code}" 64
}

test_rotate_when_there_are_no_logs_it_should_say_so() {
    assert_contains "$("${BIN}/beacon-rotate" --dir "${SCRATCH}")" "nothing to rotate"
}

test_every_script_when_asked_for_help_it_should_print_usage() {
    local script
    for script in "${BIN}"/*; do
        assert_contains "$("${script}" --help)" "Usage: $(basename "${script}")" || return 1
    done
}

passed=0 failed=0
for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
    [[ -z "${pattern}" || "${t}" == *${pattern}* ]] || continue
    SCRATCH="$(mktemp -d)"
    if ( set -e; SCRATCH="${SCRATCH}"; "${t}" ); then
        passed=$(( passed + 1 ))
        echo "ok   ${t}"
    else
        failed=$(( failed + 1 ))
        echo "FAIL ${t}"
    fi
    rm -rf "${SCRATCH}"
done

echo
echo "${passed} passed, ${failed} failed"
(( failed == 0 ))
