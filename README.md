# RasBeacon plugins

The plugins in [RasBeacon](https://github.com/aazam476/RasBeacon)'s built-in
catalog. Install them on RasBeacon from **Plugins → Browse plugins**.

| Plugin | What it does | Asks for |
|---|---|---|
| [AdGuard Home](plugins/adguard-home) | Blocks ads and trackers for every device on the hotspot; asks Quad9 over DNS-over-HTTPS. Runs AdGuard's official image, pinned by digest. | The internet |
| [Samba](plugins/samba) | Shares RasBeacon's storage, or a USB drive, with the computers on the hotspot. | A USB drive |
| [Argon ONE V2](plugins/argon-one) | The Argon ONE V2 case's fan by temperature, and its power button (Raspberry Pi 4). | I²C, GPIO, shutting down, the temperature |

How plugins work, what they may do and how to make your own:
[RasBeacon's docs/plugins.md](https://github.com/aazam476/RasBeacon/blob/main/docs/plugins.md).

## Layout

```
plugins/<id>/        one plugin: manifest.json, icon.png, README.md, and
                     templates/ or a Containerfile (and its sources)
icons/               the icons' SVG sources
tools/build.sh       builds every plugin into dist/<id>.rba (and the catalog)
tools/catalog.py     writes dist/rasbeacon-catalog.json
```

## Building

On an arm64 Linux machine with Docker (buildx), Go, zip and Python 3:

```sh
git clone https://github.com/aazam476/RasBeacon ../RasBeacon
go build -C ../RasBeacon -o /tmp/rbsign ./cmd/rbsign
tools/build.sh /tmp/rbsign      # dist/*.rba, not signed
```

Install a file from `dist/` with **Plugins → Install from a file**.

## Releases and signing

The Release workflow builds every plugin, signs it and publishes a GitHub
release; RasBeacon's catalog reads the newest one. Push a tag
(`git tag v2026.10.1 && git push origin v2026.10.1`) or run the workflow
from the Actions tab.

The plugins are signed with the key in the repository secret
`RASBEACON_PLUGIN_SIGNING_KEY` (make one with `rbsign plugin-keygen`; it is
never committed). RasBeacon remembers the key at the first install and only
accepts updates signed with it, so **keep the key safe and don't replace
it**: owners would have to remove and reinstall every plugin.

## License

Apache-2.0, like RasBeacon (see [LICENSE](LICENSE)). AdGuard Home and Samba
are their authors' programs under their own licenses (GPL-3.0); the plugins
here only package and configure them.
