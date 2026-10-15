#!/usr/bin/env bash
# Builds every plugin in plugins/ into dist/<id>.rba, with its icon and
# dist/rasbeacon-catalog.json for RasBeacon's catalog.
#
# Usage: tools/build.sh [RBSIGN]
#   RBSIGN is RasBeacon's rbsign program; without it, tools/rbsign.sh builds
#   one. With RASBEACON_PLUGIN_SIGNING_KEY set, the plugins are signed;
#   without it they are not (for trying them out and checking pull
#   requests).
#
# Needs docker with buildx (images are built for linux/arm64; an arm64
# machine is fastest), zip and python3.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
if [ $# -ge 1 ]; then
  rbsign="$(realpath "$1")"
else
  rbsign="$root/tools/rbsign.sh"
fi
dist="$root/dist"
rm -rf "$dist"
mkdir -p "$dist"

for dir in "$root"/plugins/*/; do
  id="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["id"])' "$dir/manifest.json")"
  version="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["version"])' "$dir/manifest.json")"
  echo "== $id $version"
  stage="$(mktemp -d)"
  cp "$dir/manifest.json" "$dir/icon.png" "$stage/"
  [ -f "$dir/README.md" ] && cp "$dir/README.md" "$stage/"
  [ -d "$dir/templates" ] && cp -r "$dir/templates" "$stage/"

  if [ -f "$dir/Containerfile" ]; then
    # Its own image, carried in the file.
    tag="localhost/rasbeacon-plugin-build/$id:$version"
    docker buildx build --platform linux/arm64 --file "$dir/Containerfile" --tag "$tag" --load "$dir"
    docker save "$tag" --output "$stage/image.tar"
  fi

  # Fixed timestamps and order, so the same sources give the same file.
  find "$stage" -exec touch -d 2020-01-01T00:00:00Z {} +
  (cd "$stage" && find . -type f | sed 's|^\./||' | LC_ALL=C sort | zip -X -D -q "$dist/$id.unsigned.rba" -@)
  if [ -n "${RASBEACON_PLUGIN_SIGNING_KEY:-}" ]; then
    "$rbsign" plugin-sign -in "$dist/$id.unsigned.rba" -out "$dist/$id.rba"
    rm "$dist/$id.unsigned.rba"
  else
    mv "$dist/$id.unsigned.rba" "$dist/$id.rba"
  fi
  "$rbsign" plugin-check "$dist/$id.rba"
  cp "$dir/icon.png" "$dist/$id.png"
  rm -rf "$stage"
done

python3 "$root/tools/catalog.py" "$root/plugins" "$dist"
