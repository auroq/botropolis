#!/usr/bin/env bash
# Printed once after install. Botropolis needs three things switched on that a
# package cannot switch on for you: a user daemon, the hooks, and the shell
# integration that makes a bare `claude` a background session.
set -e

cat <<'HINT'

Botropolis is installed. Three steps to finish:

  systemctl --user enable --now botropolisd
  botropolis install-hooks
  echo 'source /usr/share/botropolis/botropolis.bash' >> ~/.bashrc

Then `botropolis` opens the city, and `botropolis doctor` says what is missing.

HINT
