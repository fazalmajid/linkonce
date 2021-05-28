# linkonce - sync directories while allowing deletions on the destination

`linkonce -d $DESTDIR` will recreate the current directory's structure in `$DESTDIR` (using hard links so no extra disk space is taken) and remember which files have been linked in the file `.linkonce` so it does not attempt to link them again.

I use this to maintain a copy of the photos taken by my wife and backed up using rsync and the [Photobackup](https://apps.apple.com/us/app/photobackup-backup-photos-and-videos-via-rsync/id945026388) app, and import them into Lightroom, but I also want to curate them. Thus new photos will be linked into the curated directory `$DESTDIR`, but ones that I deleted will stay deleted, and not reappear each time rsync kicks in. This does mean changes to the files will not be reflected either as subsequent rsync runs on changed files will break the hard link.

## Dependencies

Go 1.16 or later

## Building

```
go get github.com/fazalmajid/linkonce
```
