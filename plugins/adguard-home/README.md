AdGuard Home answers your devices' DNS questions and blocks the ones for ads,
trackers and malware. It asks **Quad9** over DNS-over-HTTPS for everything
else, through the VPN when it is on.

**After installing:** on RasBeacon's Internet page, set the DNS servers to
the address shown under *Next steps*, and turn on **Use these servers while
the VPN is on**. Guests' lookups are filtered too, but only your devices can
open AdGuard Home's web interface.

In AdGuard Home all questions come from RasBeacon (it asks on your devices'
behalf), so its per-device lists show one client.

AdGuard Home is made by AdGuard (GPL-3.0); this plugin runs their official
image, pinned to one version.
