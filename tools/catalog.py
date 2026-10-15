"""Writes dist/rasbeacon-catalog.json: what RasBeacon's catalog shows for
each plugin without downloading it (see RasBeacon's docs/plugins.md).

For a plugin whose image comes from a registry, downloadSize is the
compressed size of its arm64 image, asked with `docker buildx imagetools`.

Usage: python3 tools/catalog.py PLUGINS_DIR DIST_DIR
"""

import json
import pathlib
import subprocess
import sys


def registry_size(image):
    """The arm64 image's download size, in bytes (0 if unknown)."""
    ref = image["ref"].rsplit(":", 1)[0] + "@" + image["digest"]
    try:
        raw = subprocess.run(
            ["docker", "buildx", "imagetools", "inspect", "--raw", ref],
            check=True, capture_output=True, text=True,
        ).stdout
        manifest = json.loads(raw)
        if "manifests" in manifest:
            arm = next(m for m in manifest["manifests"]
                       if m.get("platform", {}).get("architecture") == "arm64")
            raw = subprocess.run(
                ["docker", "buildx", "imagetools", "inspect", "--raw", ref.split("@")[0] + "@" + arm["digest"]],
                check=True, capture_output=True, text=True,
            ).stdout
            manifest = json.loads(raw)
        return manifest["config"]["size"] + sum(layer["size"] for layer in manifest["layers"])
    except (subprocess.CalledProcessError, StopIteration, KeyError, json.JSONDecodeError) as problem:
        print(f"warning: no download size for {ref}: {problem}", file=sys.stderr)
        return 0


def main():
    plugins, dist = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
    entries = []
    for manifest_file in sorted(plugins.glob("*/manifest.json")):
        m = json.loads(manifest_file.read_text())
        entry = {
            "file": f"{m['id']}.rba",
            "id": m["id"],
            "name": m["name"],
            "version": m["version"],
            "description": m.get("description", ""),
            "permissions": m.get("permissions", []),
            "boards": m.get("boards", []),
            "icon": f"{m['id']}.png",
        }
        if "image" in m:
            entry["downloadSize"] = registry_size(m["image"])
        entries.append(entry)
    (dist / "rasbeacon-catalog.json").write_text(json.dumps({"plugins": entries}, indent=2) + "\n")
    print(f"catalog: {len(entries)} plugins")


if __name__ == "__main__":
    main()
