# The demo is filmed in containers so that nothing it touches can be the
# real home: no ~/.claude, no ~/.claude.json, no XDG state, no network.
#
#   renderer  runs botropolis headless and encodes the clips.
#   recorder  runs Claude Code against the toy projects to make the corpus.
#
# Binaries are built on the host (CGO_ENABLED=0, see the Makefile) and
# mounted in, so the image never needs rebuilding for a code change.

FROM debian:trixie-slim AS renderer
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      xvfb xauth \
      libgl1 libegl1 libgl1-mesa-dri \
      libx11-6 libxcursor1 libxi6 libxinerama1 libxrandr2 \
      ffmpeg tzdata \
 && rm -rf /var/lib/apt/lists/*
ENV LIBGL_ALWAYS_SOFTWARE=1
WORKDIR /work
