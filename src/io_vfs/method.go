package io_vfs

import (
	"bytes"
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/avfs/avfs"

	"acme-deploy/src/io_fs"
	"acme-deploy/src/l"
)

func (receiver *VFSDB) MustReadlink(name string) string {
	switch outbound, err := receiver.VFS.Readlink(name); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"name": name} */)
		return ""
	default:
		return outbound
	}
}

func (receiver *VFSDB) MustReadFile(name string) []byte {
	switch outbound, err := receiver.VFS.ReadFile(name); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"name": name} */)
		return nil
	default:
		return outbound
	}
}

func (receiver *VFSDB) MustWriteFile(filename string, data []byte) {
	switch err := receiver.VFS.WriteFile(filename, data, avfs.DefaultFilePerm); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"name": filename} */)
	}
}

func (receiver *VFSDB) MustGetFullAbs(listID string, name string) string {
	switch {
	case len(listID) == 0:
		return receiver.MustAbs(name)
	}

	switch value, ok := receiver.List[listID]; {
	case ok:
		return receiver.MustAbs(filepath.Join(value, name))
	default:
		l.Critical.E(ErrListIDNotFound, l.F{"list ID": listID})
		return ""
	}
}

func (receiver *VFSDB) MustLReadFile(listID string, name string) []byte {
	return receiver.MustReadFile(receiver.MustGetFullAbs(listID, name))
}

func (receiver *VFSDB) MustLWriteFile(listID string, filename string, data []byte) {
	receiver.MustWriteFile(receiver.MustGetFullAbs(listID, filename), data)
}

func (receiver *VFSDB) MustMkdirAll(path string) {
	switch err := receiver.VFS.MkdirAll(path, avfs.DefaultDirPerm); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"name": path} */)
	}
}

func (receiver *VFSDB) MustSymlink(oldname string, newname string) {
	switch err := receiver.VFS.Symlink(oldname, newname); {
	case err != nil:
		l.Critical.E(err, nil /*l.F{"oldname": oldname, "newname": newname}*/)
	}
}

func (receiver *VFSDB) MustReadVFS() {
	for a, b := range receiver.List {

		receiver.List[a] = io_fs.MustAbs(b)

		io_fs.MustWalkDir(receiver.List[a], func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Critical.E(err, l.F{"name": name})
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				receiver.MustMkdirAll(name)

			case fs.ModeSymlink:
				var (
					target = io_fs.MustReadLink(name)
				)

				receiver.MustSymlink(target, name)

				// FIXME implement fs.DirEntry Type() check in case of symlink->symlink / symlink->dir, etc ....
				receiver.MustMkdirAll(io_fs.Dir(target))
				receiver.MustWriteFile(target, io_fs.MustReadFile(name))

			case 0:
				receiver.MustWriteFile(name, io_fs.MustReadFile(name))

			default:
			}

			return nil
		})
	}
}

func (receiver *VFSDB) MustWriteVFS() {
	// remove described-only orphaned entries from FS
	var (
		orphanList = make(map[string]bool)
		orphanFn   = func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Critical.E(err, l.F{"name": name})
			}

			switch orphanFileInfo, orphanErr := receiver.VFS.Lstat(name); {
			case errors.Is(orphanErr, fs.ErrNotExist): //							not exist
				orphanList[name] = true
			case orphanErr != nil: //												error
				l.Critical.E(err, nil /* l.F{"name": name} */)
			case dirEntry.Type() != orphanFileInfo.Mode().Type(): //				exist but different type
				orphanList[name] = true

			case dirEntry.Type() == fs.ModeSymlink &&
				dirEntry.Type() == orphanFileInfo.Mode().Type() &&
				io_fs.MustReadLink(name) != receiver.MustReadlink(name): // check symlink match
				orphanList[name] = true
			}

			return nil
		}
	)

	for _, b := range receiver.List {
		receiver.MustWalkDir(b, orphanFn)
	}

	for a, b := range orphanList {
		switch b {
		case true:
			l.Notice.E(ErrOrphanedEntry, l.F{"name": a})
			// io_fs.MustRemove(a)
		}
	}

	// compare and sync VFS to FS
	var (
		syncFn = func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Critical.E(err, l.F{"name": name})
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				io_fs.MustMkdir(name)
			case fs.ModeSymlink:
				io_fs.MustSymlink(receiver.MustReadlink(name), name)
			case 0:
				switch {
				case io_fs.MustIsNotExist(name) || !bytes.Equal(io_fs.MustReadFile(name), receiver.MustReadFile(name)):
					io_fs.MustWriteFile(name, receiver.MustReadFile(name))
				}
			default:
			}

			return nil
		}
	)

	receiver.MustWalkDir("/", syncFn)
}

func (receiver *VFSDB) MustWalkDir(root string, fn fs.WalkDirFunc) {
	switch err := receiver.VFS.WalkDir(root, fn); {
	case err != nil:
		l.Critical.E(err, l.F{"name": root})
	}
}
func (receiver *VFSDB) MustAbs(path string) string {
	switch outbound, err := receiver.VFS.Abs(path); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"name": path} */)
		return ""
	default:
		return outbound
	}
}

func (receiver *VFSDB) MustGlob(pattern string) []string {
	switch outbound, err := receiver.VFS.Glob(pattern); {
	case err != nil:
		l.Critical.E(err, nil /* l.F{"pattern": pattern} */)
		return nil
	default:
		return outbound
	}
}
func (receiver *VFSDB) MustLGlob(listID string, pattern string) []string {
	return receiver.MustGlob(receiver.MustGetFullAbs(listID, pattern))
}
