Shares files with the Windows, macOS and Linux computers on your hotspot.

Pick a **USB drive** to share it (exFAT, FAT32, NTFS or ext4), or
**RasBeacon internal storage** to keep the files on RasBeacon's own SD card.
Nothing on the drive can run on RasBeacon.

- **Mac:** Finder → Go → Connect to Server, `smb://<RasBeacon's address>`.
- **Windows:** File Explorer, `\\<RasBeacon's address>` (Windows needs port
  445).

**How much room:** on RasBeacon's internal storage, you choose the size when
installing (2 GB to start), and can make it bigger or smaller later on its
page, down to what the files use. That space is set aside on the SD card.
With a USB drive, the **Size limit** setting (optional, shown only with a
drive) makes the share look that big to the computers: Finder and File
Explorer don't copy more onto it. A limit below what the files already use
deletes nothing; the share shows as full until files are removed.

Saving a change restarts Samba, so a copy in progress stops, and the
computer has to copy that file again. Change sizes when nothing is copying.

Guests can't reach it. Needs RasBeacon 0.3.0 or newer.
