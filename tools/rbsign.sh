#!/usr/bin/env bash
# Runs rbsign, RasBeacon's tool for plugin keys, signatures and checks,
# building it first if needed. rbsign is part of RasBeacon
# (github.com/aazam476/RasBeacon, cmd/rbsign), so this fetches RasBeacon's
# source into .cache/ the first time.
#
# Usage:
#   tools/rbsign.sh plugin-keygen                  make a signing key (once)
#   tools/rbsign.sh plugin-check dist/*.rba        check plugin files
#   RASBEACON_PLUGIN_SIGNING_KEY=... tools/rbsign.sh plugin-sign -in a.rba -out b.rba
#
# Needs git and Go 1.24 or newer. RASBEACON_REF picks RasBeacon's branch or
# tag (default main).
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cache="$root/.cache"
source_dir="$cache/RasBeacon"
binary="$cache/rbsign"
ref="${RASBEACON_REF:-main}"

if [ ! -x "$binary" ] || [ "${RBSIGN_REBUILD:-}" = 1 ]; then
  mkdir -p "$cache"
  if [ -d "$source_dir/.git" ]; then
    git -C "$source_dir" fetch --quiet --depth 1 origin "$ref"
    git -C "$source_dir" checkout --quiet FETCH_HEAD
  else
    git clone --quiet --depth 1 --branch "$ref" https://github.com/aazam476/RasBeacon "$source_dir"
  fi
  echo "Building rbsign from RasBeacon ($ref)..." >&2
  go build -C "$source_dir" -o "$binary" ./cmd/rbsign
fi

exec "$binary" "$@"
