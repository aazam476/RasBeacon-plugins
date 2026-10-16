# <img src=".github/logo.svg" alt="" height="32"> RasBeacon plugins

The plugins in [RasBeacon](https://github.com/aazam476/RasBeacon)'s built-in
catalog. Install them on RasBeacon from **Plugins → Browse plugins**.

| Plugin | What it does | Asks for |
|---|---|---|
| [AdGuard Home](plugins/adguard-home) | Blocks ads and trackers for every device on the hotspot; asks Quad9 over DNS-over-HTTPS. Runs AdGuard's official image, pinned by digest. | The internet |
| [Samba](plugins/samba) | Shares RasBeacon's storage, or a USB drive, with the computers on the hotspot. | A USB drive |
| [Argon ONE V2](plugins/argon-one) | The Argon ONE V2 case's fan by temperature, and its power button (Raspberry Pi 4). | I²C, GPIO, shutting down, the temperature |

**Making your own plugins?** Start from
[RasBeacon-plugin-template](https://github.com/aazam476/RasBeacon-plugin-template):
a small working plugin, the same tools as this repository, and a guide from
hello world to plugins like these. The reference for every field is
[RasBeacon's docs/plugins.md](https://github.com/aazam476/RasBeacon/blob/main/docs/plugins.md).

## Layout

```
plugins/<id>/        one plugin: manifest.json, icon.png, README.md, and
                     templates/ or a Containerfile (and its sources)
icons/               the icons' SVG sources
tools/build.sh       builds every plugin into dist/<id>.rba (and the catalog)
tools/catalog.py     writes dist/rasbeacon-catalog.json
tools/rbsign.sh      RasBeacon's rbsign: signing keys, signing, checking
```

## Building

On Linux with Docker (buildx), Go 1.24 or newer, git, zip and Python 3 (an
arm64 machine is fastest; the images are for the Raspberry Pi):

```sh
tools/build.sh      # dist/*.rba, not signed
```

Install a file from `dist/` with **Plugins → Install from a file**.

`tools/rbsign.sh` builds `rbsign` from
[RasBeacon's source](https://github.com/aazam476/RasBeacon/tree/main/cmd/rbsign)
the first time (into `.cache/`). Check any plugin file the way RasBeacon
does with `tools/rbsign.sh plugin-check dist/*.rba`.

## Signing key (once)

Releases are signed, so RasBeacon can tell that an update comes from the
same author (see [Signing](https://github.com/aazam476/RasBeacon/blob/main/docs/plugins.md#signing)).
Make the key once:

1. Make a key on your computer:

   ```sh
   tools/rbsign.sh plugin-keygen
   ```

   It prints a **private key**, a public key and a key ID (RasBeacon shows the
   ID when installing your plugins).
2. Save the private key somewhere safe, such as a password manager. Never
   commit it or paste it into an issue or chat.
3. On GitHub, open this repository's **Settings → Secrets and variables →
   Actions → New repository secret**, name it
   `RASBEACON_PLUGIN_SIGNING_KEY` and paste the private key.

Keep the same key from then on. RasBeacon remembers it at the first install
and accepts only updates signed with it; with a new key, owners would have to
remove and reinstall every plugin.

## Publishing a release

RasBeacon's catalog lists the plugins in this repository's **newest
release**. To publish one, open **Actions → Release → Run workflow**, type a
release name, and run it (or push a tag with that name).

Release names follow RasBeacon's version: `v0.1.0`, then `v0.1.1`, `v0.1.2`,
… are plugins for RasBeacon 0.1; `v0.2.0` starts the plugins for RasBeacon
0.2. A plugin that needs a newer RasBeacon than an owner has says so with
`minRasBeacon` in its manifest, and older RasBeacons skip it.

> These plugins are new and not yet fully tested on a Raspberry Pi, like
> RasBeacon 0.1 itself. Please report problems as issues.
The workflow builds every plugin, signs it with the key, and publishes the
release with the `.rba` files, their icons and `rasbeacon-catalog.json`.

Bump a plugin's `version` in its `manifest.json` whenever it changes:
RasBeacon offers an update only for a newer version.

## License

Apache-2.0, like RasBeacon (see [LICENSE](LICENSE)). AdGuard Home and Samba
are their authors' programs under their own licenses (GPL-3.0); the plugins
here only package and configure them.
