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
		fn = func(name string, dirEntry fs.DirEntry, fnErr error) (err error) {
			switch {
			case fnErr != nil:
				return fnErr
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				switch err = r.VFS.MkdirAll(name, avfs.DefaultDirPerm); {
				case err != nil:
					return
				}

			case fs.ModeSymlink:
				var (
					target string
					data   []byte
				)

				switch target, err = os.Readlink(name); {
				case err != nil:
					return
				}

				switch err = r.VFS.Symlink(target, name); {
				case err != nil:
					return
				}

				// FIXME implement fs.DirEntry Type() check in case of symlink->symlink / symlink->dir, etc ....
				switch err = r.VFS.MkdirAll(io_fs.Dir(target), avfs.DefaultDirPerm); {
				case err != nil:
					return
				}

				switch data, err = os.ReadFile(name); {
				case err != nil:
					return
				}
				switch err = r.VFS.WriteFile(target, data, avfs.DefaultFilePerm); {
				case err != nil:
					return
				}

			case 0:
				switch err = r.VFS.MkdirAll(io_fs.Dir(name), avfs.DefaultDirPerm); {
				case err != nil:
					return
				}
				switch err = r.CopyFileFS2VFS(name); {
				case err != nil:
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

func (r *VFSDB) MustWriteVFS() (err error) {
	// remove described-only orphaned entries from FS
	var (
		orphanList = make(map[string]struct{})
		orphanFn   = func(name string, dirEntry fs.DirEntry, fnErr error) (err error) {
			switch {
			case fnErr != nil:
				return fnErr
			}

			var (
				orphanFileInfo fs.FileInfo
			)
			switch orphanFileInfo, err = r.VFS.Lstat(name); {
			case errors.Is(err, fs.ErrNotExist): //							not exist
				orphanList[name] = struct{}{}
			case err != nil: //												error
				return err

			case dirEntry.Type() != orphanFileInfo.Mode().Type(): //				exist but different type
				orphanList[name] = struct{}{}

			case dirEntry.Type() == fs.ModeSymlink &&
				dirEntry.Type() == orphanFileInfo.Mode().Type() &&
				io_fs.MustReadLink(name) != r.MustReadlink(name): //         check symlink match
				orphanList[name] = struct{}{}
			}

			return
		}
	)

	for _, b := range r.List {
		switch err = r.VFS.WalkDir(b, orphanFn); {
		case err != nil:
			return
		}
	}

	for a := range orphanList {
		l.Z{l.E: l.EORPHANED, "name": a}.Notice()
	}

	// compare and sync VFS to FS
	var (
		syncFn = func(name string, dirEntry fs.DirEntry, fnErr error) (err error) {
			switch {
			case fnErr != nil:
				return fnErr
			}

			switch dirEntry.Type() {
			case fs.ModeDir:
				switch err = os.Mkdir(name, avfs.DefaultDirPerm); {
				case errors.Is(err, fs.ErrExist):
				case err != nil:
					return
				}
			case fs.ModeSymlink:
				io_fs.MustSymlink(r.MustReadlink(name), name)
			case 0:
				switch {
				case io_fs.MustIsNotExist(name) || !bytes.Equal(io_fs.MustReadFile(name), r.MustReadFile(name)):
					io_fs.MustWriteFile(name, r.MustReadFile(name))
				}
			default:
			}

			return
		}
	)

	switch err = r.VFS.WalkDir("/", syncFn); {
	case err != nil:
		return
	}

	return
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
