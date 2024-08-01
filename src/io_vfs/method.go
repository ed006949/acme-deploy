package io_vfs

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
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
		l.Critical.E(err, l.F{"name": name})
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
		l.Critical.E(l.ENOTFOUND, l.F{"list ID": listID})
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
func (receiver *VFSDB) MustCopyFS2VFS() {
	switch err := receiver.CopyFS2VFS(); {
	case err != nil:
		l.Critical.E(err, nil)
	}
}

func (receiver *VFSDB) CopyFS2VFS() (err error) {
	for a, b := range receiver.List {
		switch receiver.List[a], err = filepath.Abs(b); {
		case err != nil:
			return
		}

		switch err = receiver.CopyFromFS2VFS(receiver.List[a]); {
		case err != nil:
			return
		}
	}
	return
}

func (receiver *VFSDB) CopyFromFS2VFS(name string) (err error) {
	var (
		fn = func(name string, dirEntry fs.DirEntry, err error) (fnErr error) {
			switch {
			case err != nil:
				return err
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				switch fnErr = receiver.VFS.MkdirAll(name, avfs.DefaultDirPerm); {
				case fnErr != nil:
					return
				}

			case fs.ModeSymlink:
				var (
					target string
					data   []byte
				)

				switch target, fnErr = os.Readlink(name); {
				case fnErr != nil:
					return
				}

				switch fnErr = receiver.VFS.Symlink(target, name); {
				case fnErr != nil:
					return
				}

				// FIXME implement fs.DirEntry Type() check in case of symlink->symlink / symlink->dir, etc ....
				switch fnErr = receiver.VFS.MkdirAll(io_fs.Dir(target), avfs.DefaultDirPerm); {
				case fnErr != nil:
					return
				}

				switch data, fnErr = os.ReadFile(name); {
				case fnErr != nil:
					return
				}
				switch fnErr = receiver.VFS.WriteFile(target, data, avfs.DefaultFilePerm); {
				case fnErr != nil:
					return
				}

			case 0:
				switch fnErr = receiver.VFS.MkdirAll(io_fs.Dir(name), avfs.DefaultDirPerm); {
				case fnErr != nil:
					return
				}
				switch fnErr = receiver.CopyFileFS2VFS(name); {
				case fnErr != nil:
					return
				}

			default:
			}

			return
		}
	)

	switch name, err = filepath.Abs(name); {
	case err != nil:
		return
	}

	switch err = filepath.WalkDir(name, fn); {
	case err != nil:
		return
	}

	return
}

func (receiver *VFSDB) MustWriteVFS() {
	// remove described-only orphaned entries from FS
	var (
		orphanList = make(map[string]struct{})
		orphanFn   = func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Critical.E(err, l.F{"name": name})
			}

			switch orphanFileInfo, orphanErr := receiver.VFS.Lstat(name); {
			case errors.Is(orphanErr, fs.ErrNotExist): //							not exist
				orphanList[name] = struct{}{}
			case orphanErr != nil: //												error
				l.Critical.E(err, nil /* l.F{"name": name} */)

			case dirEntry.Type() != orphanFileInfo.Mode().Type(): //				exist but different type
				orphanList[name] = struct{}{}

			case dirEntry.Type() == fs.ModeSymlink &&
				dirEntry.Type() == orphanFileInfo.Mode().Type() &&
				io_fs.MustReadLink(name) != receiver.MustReadlink(name): //         check symlink match
				orphanList[name] = struct{}{}
			}

			return nil
		}
	)

	for _, b := range receiver.List {
		receiver.MustWalkDir(b, orphanFn)
	}

	for a := range orphanList {
		l.Notice.E(l.EORPHANED, l.F{"name": a})
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

func (receiver *VFSDB) CopyFileFS2VFS(name string) (err error) {
	var (
		data []byte
	)

	// switch name, err = filepath.Abs(name); {
	// case err != nil:
	// 	return
	// }
	switch data, err = os.ReadFile(name); {
	case err != nil:
		return
	}
	switch err = receiver.VFS.WriteFile(name, data, avfs.DefaultFilePerm); {
	case err != nil:
		return
	}
	return
}
