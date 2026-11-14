#!/bin/sh
# Starts Samba with one share, /share (RasBeacon mounts the chosen USB drive
# there, or the plugin's own storage), for one user.
set -eu

: "${SMB_SHARE:=Files}"
: "${SMB_USER:=traveler}"
: "${SMB_PASSWORD:?the password setting is empty}"
# The size limit in GB, if the owner set one (empty: none).
: "${SMB_SIZE_LIMIT_GB:=}"

case "$SMB_SHARE" in
  *[!A-Za-z0-9\ _-]*|"") echo "the share name may only have letters, digits, spaces, - and _" >&2; exit 1 ;;
esac
case "$SMB_USER" in
  *[!a-z0-9_-]*|"") echo "the username may only have lowercase letters, digits, - and _" >&2; exit 1 ;;
esac

# The size limit, for dfree.sh, in 1 KiB blocks. RasBeacon already checked it
# is a number from 1 to 16384; awk turns GB with decimals into whole blocks.
if [ -n "$SMB_SIZE_LIMIT_GB" ]; then
  awk -v gb="$SMB_SIZE_LIMIT_GB" 'BEGIN { printf "%d\n", gb * 1024 * 1024 }' > /etc/samba/size-limit-kb
else
  rm -f /etc/samba/size-limit-kb
fi

mkdir -p /share /var/lib/samba/private /run/samba
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
  # Files on the drive belong to the container's root; everyone who logs in
  # writes as the same user.
  force user = root
  # The share's size and free space (with the size limit, if set). Asked
  # again at most once a minute.
  dfree command = /usr/local/bin/dfree
  dfree cache time = 60

[$SMB_SHARE]
  path = /share
  valid users = $SMB_USER
  read only = no
  browseable = yes
  # macOS: no ._ files and .DS_Store clutter.
  vfs objects = catia fruit streams_xattr
  fruit:metadata = stream
  veto files = /._*/.DS_Store/
  delete veto files = yes
CONF

exec smbd --foreground --no-process-group --debug-stdout
