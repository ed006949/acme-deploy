package io_vfs

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/avfs/avfs"
	"github.com/go-ini/ini"

	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_fs"
	"acme-deploy/src/l"
)

func (r *VFSDB) MustReadlink(name string) string {
	switch outbound, err := r.VFS.Readlink(name); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}

func (r *VFSDB) MustReadFile(name string) []byte {
	switch outbound, err := r.VFS.ReadFile(name); {
	case err != nil:
		l.Z{l.E: err, "name": name}.Critical()
		return nil
	default:
		return outbound
	}
}

func (r *VFSDB) MustWriteFile(filename string, data []byte) {
	switch err := r.VFS.WriteFile(filename, data, avfs.DefaultFilePerm); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}

func (r *VFSDB) MustGetFullAbs(listID string, name string) string {
	switch {
	case len(listID) == 0:
		return r.MustAbs(name)
	}

	switch value, ok := r.List[listID]; {
	case ok:
		return r.MustAbs(filepath.Join(value, name))
	default:
		l.Z{l.E: l.ENOTFOUND, "list ID": listID}.Critical()
		return ""
	}
}

func (r *VFSDB) MustLReadFile(listID string, name string) []byte {
	return r.MustReadFile(r.MustGetFullAbs(listID, name))
}

func (r *VFSDB) MustLWriteFile(listID string, filename string, data []byte) {
	r.MustWriteFile(r.MustGetFullAbs(listID, filename), data)
}

func (r *VFSDB) MustMkdirAll(path string) {
	switch err := r.VFS.MkdirAll(path, avfs.DefaultDirPerm); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}

func (r *VFSDB) MustSymlink(oldname string, newname string) {
	switch err := r.VFS.Symlink(oldname, newname); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}
func (r *VFSDB) MustCopyFS2VFS() {
	switch err := r.CopyFS2VFS(); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}

func (r *VFSDB) CopyFS2VFS() (err error) {
	for a, b := range r.List {
		switch r.List[a], err = filepath.Abs(b); {
		case err != nil:
			return
		}

		switch err = r.CopyFromFS2VFS(r.List[a]); {
		case err != nil:
			return
		}
	}
	return
}

func (r *VFSDB) CopyFromFS2VFS(name string) (err error) {
	var (
		fn = func(name string, dirEntry fs.DirEntry, err error) (fnErr error) {
			switch {
			case err != nil:
				return err
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				switch fnErr = r.VFS.MkdirAll(name, avfs.DefaultDirPerm); {
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

				switch fnErr = r.VFS.Symlink(target, name); {
				case fnErr != nil:
					return
				}

				// FIXME implement fs.DirEntry Type() check in case of symlink->symlink / symlink->dir, etc ....
				switch fnErr = r.VFS.MkdirAll(io_fs.Dir(target), avfs.DefaultDirPerm); {
				case fnErr != nil:
					return
				}

				switch data, fnErr = os.ReadFile(name); {
				case fnErr != nil:
					return
				}
				switch fnErr = r.VFS.WriteFile(target, data, avfs.DefaultFilePerm); {
				case fnErr != nil:
					return
				}

			case 0:
				switch fnErr = r.VFS.MkdirAll(io_fs.Dir(name), avfs.DefaultDirPerm); {
				case fnErr != nil:
					return
				}
				switch fnErr = r.CopyFileFS2VFS(name); {
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

func (r *VFSDB) MustWriteVFS() {
	// remove described-only orphaned entries from FS
	var (
		orphanList = make(map[string]struct{})
		orphanFn   = func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Z{l.E: err, "name": name}.Critical()
			}

			switch orphanFileInfo, orphanErr := r.VFS.Lstat(name); {
			case errors.Is(orphanErr, fs.ErrNotExist): //							not exist
				orphanList[name] = struct{}{}
			case orphanErr != nil: //												error
				l.Z{l.E: err}.Critical()

			case dirEntry.Type() != orphanFileInfo.Mode().Type(): //				exist but different type
				orphanList[name] = struct{}{}

			case dirEntry.Type() == fs.ModeSymlink &&
				dirEntry.Type() == orphanFileInfo.Mode().Type() &&
				io_fs.MustReadLink(name) != r.MustReadlink(name): //         check symlink match
				orphanList[name] = struct{}{}
			}

			return nil
		}
	)

	for _, b := range r.List {
		r.MustWalkDir(b, orphanFn)
	}

	for a := range orphanList {
		l.Z{l.E: l.EORPHANED, "name": a}.Notice()
	}

	// compare and sync VFS to FS
	var (
		syncFn = func(name string, dirEntry fs.DirEntry, err error) error {
			switch {
			case err != nil:
				l.Z{l.E: err, "name": name}.Critical()
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				io_fs.MustMkdir(name)
			case fs.ModeSymlink:
				io_fs.MustSymlink(r.MustReadlink(name), name)
			case 0:
				switch {
				case io_fs.MustIsNotExist(name) || !bytes.Equal(io_fs.MustReadFile(name), r.MustReadFile(name)):
					io_fs.MustWriteFile(name, r.MustReadFile(name))
				}
			default:
			}

			return nil
		}
	)

	r.MustWalkDir("/", syncFn)
}

func (r *VFSDB) MustWalkDir(root string, fn fs.WalkDirFunc) {
	switch err := r.VFS.WalkDir(root, fn); {
	case err != nil:
		l.Z{l.E: err, "name": root}.Critical()
	}
}
func (r *VFSDB) MustAbs(path string) string {
	switch outbound, err := r.VFS.Abs(path); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}

func (r *VFSDB) MustGlob(pattern string) []string {
	switch outbound, err := r.VFS.Glob(pattern); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return nil
	default:
		return outbound
	}
}
func (r *VFSDB) MustLGlob(listID string, pattern string) []string {
	return r.MustGlob(r.MustGetFullAbs(listID, pattern))
}

func (r *VFSDB) CopyFileFS2VFS(name string) (err error) {
	var (
		data []byte
	)

	switch data, err = os.ReadFile(name); {
	case err != nil:
		return
	}
	switch err = r.VFS.WriteFile(name, data, avfs.DefaultFilePerm); {
	case err != nil:
		return
	}
	return
}

func (r *VFSDB) LoadX509KeyPair(chain string, key string) (outbound *io_crypto.Certificate, err error) {
	var (
		chainData []byte
		keyData   []byte
	)
	switch chainData, err = r.VFS.ReadFile(chain); {
	case err != nil:
		return
	}
	switch keyData, err = r.VFS.ReadFile(key); {
	case err != nil:
		return
	}
	switch outbound, err = io_crypto.X509KeyPair(chainData, keyData); {
	case err != nil:
		return nil, err
	}
	return
}
func (r *VFSDB) LoadIniMapTo(v any, source string) (err error) {
	var (
		data []byte
	)
	switch data, err = r.VFS.ReadFile(source); {
	case err != nil:
		return
	}
	return ini.MapTo(&v, data)

	// var (
	// 	dataSet []any
	// )
	// for _, b := range source {
	// 	var (
	// 		data []byte
	// 	)
	// 	switch data, err = r.VFS.ReadFile(b); {
	// 	case err != nil:
	// 		return
	// 	}
	// 	data = bytes.ReplaceAll(
	// 		data,
	// 		[]byte("/var/etc/acme-client/"),
	// 		[]byte(r.List["acme-client"]+"/"),
	// 	)
	// 	dataSet = append(dataSet, data)
	// }
	//
	// switch len(dataSet) {
	// case 0:
	// case 1:
	// 	err = ini.MapTo(v, dataSet[0])
	// default:
	// 	err = ini.MapTo(v, dataSet[0], dataSet[1:]...)
	// }
	//
	// return
}
