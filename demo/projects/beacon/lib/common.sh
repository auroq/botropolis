# shellcheck shell=bash

BEACON_ROOT="$(cd -- "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
BEACON_STATUS_FILE="${BEACON_STATUS_FILE:-${BEACON_ROOT}/data/status.txt}"
BEACON_LOG_DIR="${BEACON_LOG_DIR:-${BEACON_ROOT}/data/logs}"

die() {
    printf '%s: %s\n' "$(basename "$0")" "$*" >&2
    exit 64
}

now_epoch() {
    if [[ -n "${BEACON_NOW:-}" ]]; then
        date -u -d "${BEACON_NOW}" +%s
    else
        date -u +%s
    fi
}

iso_to_epoch() {
    date -u -d "$1" +%s 2>/dev/null || die "bad timestamp: $1"
}

require_status() {
    [[ -r "$1" ]] || die "cannot read status file: $1"
}

read_status() {
    local file="$1"
    grep -v -e '^[[:space:]]*#' -e '^[[:space:]]*$' "${file}"
}
