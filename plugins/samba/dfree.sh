#!/bin/sh
# Samba's "dfree command": how big a share is, and how much of it is free,
# as the computers on the hotspot see it. Samba runs it in the share's
# folder (with a folder in it as the argument) and reads one line: the
# total and the free space, in 1 KiB blocks.
#
# A share looks as big as its quota (written by entrypoint.sh in
# /etc/samba/quotas/<id>), and only what is left of it after its files is
# free, or what the disk has, if that is less. Once its files reach the
# quota (or pass it: files added on a computer while a USB drive was out),
# nothing is free, so Finder and File Explorer copy nothing more onto it,
# while deleting still works.
set -eu

# The share is the folder under /shares that the asked-about folder is in.
folder=$(cd "${1:-.}" && pwd -P)
id=$(echo "$folder" | sed -n 's#^/shares/\([^/]*\).*#\1#p')

# df -kP prints a header, then: filesystem, total, used, free (1 KiB blocks).
set -- $(df -kP "$folder" | awk 'NR == 2 { print $2, $3, $4 }')
total=$1 free=$3

if [ -n "$id" ] && [ -s "/etc/samba/quotas/$id" ]; then
  quota=$(cat "/etc/samba/quotas/$id")
  # The share's own files (on a USB drive, other folders share the disk).
  used=$(du -sk "/shares/$id" | cut -f1)
  left=$((quota > used ? quota - used : 0))
  total=$quota
  free=$((free < left ? free : left))
fi

echo "$total $free"
