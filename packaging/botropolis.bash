# Source from ~/.bashrc:  . /usr/share/botropolis/botropolis.bash
#
# Makes a plain `claude` (or `claude "prompt"`) from an interactive shell start
# the session in the background and attach to it, so closing the terminal
# parks the session instead of killing it.  Anything that starts with a flag
# or is a subcommand (`claude attach`, `claude --resume`, `claude agents`, ...)
# goes straight to the real CLI.

claude() {
    if [ -t 0 ] && [ -t 1 ] && { [ $# -eq 0 ] || [ "${1#-}" = "$1" ] && ! _botropolis_is_subcommand "$1"; }; then
        local id
        id=$(botropolis new "$PWD" "$@") || return $?
        command claude attach "$id"
    else
        command claude "$@"
    fi
}

_botropolis_is_subcommand() {
    case "$1" in
        agents|attach|logs|stop|rm|mcp|plugin|config|update|doctor|auth|login|logout|install|setup-token|remote-control) return 0 ;;
        *) return 1 ;;
    esac
}
