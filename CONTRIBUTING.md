# Contributing to RasBeacon plugins

Thanks for helping! Like RasBeacon itself, everything here is written to be
readable by a junior engineer or a second- or third-year computer science
student. If something is hard to follow, that is a bug worth reporting.

## Adding or changing a plugin

1. Read [RasBeacon's docs/plugins.md](https://github.com/aazam476/RasBeacon/blob/main/docs/plugins.md):
   the `.rba` format, the manifest, permissions and placeholders.
2. Make a folder `plugins/<id>/` with `manifest.json`, `icon.png` (square,
   256×256; keep the SVG source in `icons/`) and `README.md`, plus either a
   `Containerfile` (an image built here) or `image` in the manifest (a
   published image, pinned by digest).
3. Ask for as few permissions as the plugin needs. A plugin that only answers
   on its ports needs none.
4. Build it (`tools/build.sh`, see README.md), install the file on a
   RasBeacon with **Install from a file**, and go through what it does.
5. Bump `version` whenever anything in the plugin changes: RasBeacon offers an
   update only for a newer version.

Images must be for `linux/arm64` and should be small: RasBeacon may run on a
Raspberry Pi Zero 2 W with 512 MB of memory. Pin base images and published
images to a version (and published images to a digest).

## Everyday commands

```sh
cd plugins/argon-one && go test ./...    # the Argon ONE V2 program
tools/build.sh                           # every plugin into dist/ (not signed)
tools/rbsign.sh plugin-check dist/*.rba  # check files as RasBeacon does
```

CI (`.github/workflows/ci.yml`) runs the tests and builds and checks every
plugin on each push and pull request.

## Commit messages and pull requests

Short, plain-language commit messages in the imperative ("Add a Jellyfin
plugin"), one change per pull request. Say how you tested it on a Pi.

## Signing

Releases are signed by the Release workflow with the
`RASBEACON_PLUGIN_SIGNING_KEY` secret (README.md, *Signing key*, says how to
make it). Never commit a private key, and never put one on a command line.
