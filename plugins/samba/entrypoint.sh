#!/bin/sh
# Starts Samba for one user, with one Samba share for each share the owner
# made on Samba's page in RasBeacon. RasBeacon mounts each one at
# /shares/<id> and describes them all in RASBEACON_SHARES (one line of JSON:
# id, name, path, quotaBytes and settings). With no shares yet, Samba runs
# with none.
set -eu

: "${SMB_USER:=traveler}"
: "${SMB_PASSWORD:?the password setting is empty}"
: "${RASBEACON_SHARES:=[]}"

case "$SMB_USER" in
  *[!a-z0-9_-]*|"") echo "the username may only have lowercase letters, digits, - and _" >&2; exit 1 ;;
esac

mkdir -p /var/lib/samba/private /run/samba /etc/samba/quotas
id "$SMB_USER" >/dev/null 2>&1 || adduser -D -H -s /sbin/nologin "$SMB_USER"
printf '%s\n%s\n' "$SMB_PASSWORD" "$SMB_PASSWORD" | smbpasswd -s -a "$SMB_USER" >/dev/null

cat > /etc/samba/smb.conf <<CONF
[global]
  server role = standalone server
  server min protocol = SMB2_10
  map to guest = never
  load printers = no
  printing = bsd
  printcap name = /dev/null
  disable spoolss = yes
  log file = /dev/stdout
  # Files belong to the container's root; everyone who logs in writes as
  # the same user.
  force user = root
  # Each share's size and free space, by its quota (dfree.sh). Asked again
  # at most once a minute.
  dfree command = /usr/local/bin/dfree
  dfree cache time = 60
  # macOS: no ._ files and .DS_Store clutter. And no lost+found: a share
  # on RasBeacon's internal storage (or an ext4 drive) is a file system of
  # its own, whose disk check keeps that folder at its top; it isn't the
  # owner's. Hidden files can't be opened, and a folder holding only them
  # can still be deleted.
  vfs objects = catia fruit streams_xattr
  fruit:metadata = stream
  veto files = /._*/.DS_Store/lost+found/
  delete veto files = yes
CONF

# One section per share. RasBeacon already checked the names (letters,
# digits, spaces, - _ .) and the paths (its own /shares/<id>). Each quota
# goes in a file named after the share's path, for dfree.sh.
rm -f /etc/samba/quotas/*
echo "$RASBEACON_SHARES" | jq -r '.[] | [.name, .path, (.quotaBytes / 1024 | floor), (.settings.readOnly // "false")] | @tsv' |
  while IFS="$(printf '\t')" read -r name path quota_kb read_only; do
    echo "$quota_kb" > "/etc/samba/quotas/$(basename "$path")"
    read_only_value=no
    [ "$read_only" = true ] && read_only_value=yes
    cat >> /etc/samba/smb.conf <<CONF

[$name]
  path = $path
  valid users = $SMB_USER
  read only = $read_only_value
  browseable = yes
CONF
  done

exec smbd --foreground --no-process-group --debug-stdout
