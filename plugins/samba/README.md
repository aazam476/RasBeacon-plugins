Shares folders with the Windows, macOS and Linux computers on your hotspot.

Samba starts with no shares: add them on its page. Each share has a name
(what computers see), a place that never changes, and a quota:

- **On RasBeacon's internal storage**, the quota's space is set aside on the
  SD card (RasBeacon keeps 2 GB of it free). When a share is full, nothing
  more fits; deleting still works.
- **In a folder on a USB drive** (exFAT, FAT32, NTFS or ext4), nothing is set
  aside. Files already in the folder become the share's. Computers see the
  share as its quota's size, with only what its files leave of it free, and
  RasBeacon holds it to the quota: once its files reach it, nothing more can
  be added, while deleting still works. Files added on a computer while the
  drive was out count too. Nothing on a drive can run on RasBeacon.

A quota can grow up to the space there is, and shrink down to what the files
use, never below. A share can be read-only, and its files can be encrypted,
so someone who pulls out the drive or the SD card can't read them; save the
recovery key it offers.

- **Mac:** Finder → Go → Connect to Server, `smb://<RasBeacon's address>`.
- **Windows:** File Explorer, `\\<RasBeacon's address>`.

Updating from Samba 1.0 keeps your share: RasBeacon makes it the first
share, on the same USB drive, or with its files moved into a share on
internal storage.

Changing shares restarts Samba, which stops a copy in progress. Guests can't
reach it. Needs RasBeacon 0.3.0 or newer.
