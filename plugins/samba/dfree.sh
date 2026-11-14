#!/bin/sh
# Samba's "dfree command": how big the share is, and how much of it is free,
# as the computers on the hotspot see it. Samba runs it with the share's
# folder as the first argument and reads one line: the total and the free
# space, in 1 KiB blocks.
#
# With a size limit (the "Size limit" setting, written by entrypoint.sh to
# /etc/samba/size-limit-kb), the share looks that big, and only what is left
# of the limit is free (or what the drive has, if that is less). Finder and
# File Explorer don't copy more than is free, so the limit holds for them.
# It is a soft limit: Samba doesn't stop a program that writes past it.
set -eu

folder=${1:-/share}

# df -kP prints a header, then: filesystem, total, used, free (1 KiB blocks).
set -- $(df -kP "$folder" | awk 'NR == 2 { print $2, $3, $4 }')
total=$1 used=$2 free=$3

if [ -s /etc/samba/size-limit-kb ]; then
  limit=$(cat /etc/samba/size-limit-kb)
  left=$((limit > used ? limit - used : 0))
  total=$limit
  free=$((free < left ? free : left))
fi

echo "$total $free"
