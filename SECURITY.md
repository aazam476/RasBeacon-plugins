# Security policy

Plugins run on routers people trust with their traffic, so security
problems in them matter a lot to us.

## Reporting a problem

**Please don't open a public issue for security problems.** Instead, use
GitHub's private vulnerability reporting: on the repository page, go to
*Security → Report a vulnerability*. Include:

- what you found and how to reproduce it,
- which plugin and version, and which RasBeacon version and board,
- what an attacker could do with it.

We'll answer within a week, keep you updated, and credit you in the release
notes if you want.

Problems in RasBeacon's handling of plugins (a plugin doing more than its
permissions allow, signatures not being checked, guests or the internet
reaching a plugin) belong to
[RasBeacon](https://github.com/aazam476/RasBeacon/security) instead.
Problems in AdGuard Home or Samba themselves belong to their authors; we
update the plugin once they publish a fix.

## What we consider security problems

For example:

- a plugin here asking for, or using, more than it needs;
- a plugin's settings (passwords) appearing in its log or files where they
  shouldn't;
- a plugin's web interface or share reachable without its password;
- a published image that isn't pinned, or a build that doesn't use what is
  in this repository;
- a release that isn't signed with the plugins' key.

## Supported versions

Only the newest release gets security fixes. RasBeacon's catalog shows
**Update** on a plugin's page when a newer version is out.
