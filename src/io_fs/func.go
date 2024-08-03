package io_fs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/avfs/avfs"

	"acme-deploy/src/l"
)

func MustAbs(path string) string {
	switch outbound, err := filepath.Abs(path); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}
func MustGetFullAbs(path string, name string) string {
	return MustAbs(filepath.Join(path, name))
}

func MustReadLink(name string) string {
	switch outbound, err := os.Readlink(name); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}

func MustReadFile(name string) []byte {
	switch outbound, err := os.ReadFile(name); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return nil
	default:
		return outbound
	}
}

func MustWriteFile(name string, data []byte) {
	switch err := os.WriteFile(name, data, avfs.DefaultFilePerm); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}

func MustReadDir(name string) []fs.DirEntry {
	switch outbound, err := os.ReadDir(name); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return nil
	default:
		return outbound
	}
}

func MustMkdir(path string) {
	switch err := os.Mkdir(path, avfs.DefaultDirPerm); {
	case errors.Is(err, fs.ErrExist):
		l.Z{l.E: err, "name": path}.Debug()
	case err != nil:
		l.Z{l.E: err, "name": path}.Critical()
	}
}

func MustMkdirAll(path string) {
	switch err := os.MkdirAll(path, avfs.DefaultDirPerm); {
	case err != nil:
		l.Z{l.E: err, "name": path}.Critical()
	}
}

func Dir(path string) string {
	return filepath.Dir(path)
}

func MustWalkDir(root string, fn fs.WalkDirFunc) {
	switch err := filepath.WalkDir(root, fn); {
	case err != nil:
		l.Z{l.E: err, "name": root}.Critical()
	}
}

func MustIsExist(path string) bool {
	return !MustIsNotExist(path)
}

func MustIsNotExist(path string) bool {
	switch _, err := os.Stat(path); {
	case errors.Is(err, fs.ErrNotExist):
		l.Z{l.E: err, "name": path}.Debug()
		return true
	case err != nil:
		l.Z{l.E: err, "name": path}.Critical()
		return false
	default:
		return false
	}
}

func MustIsSymlink(path string) bool {
	switch stat, err := os.Lstat(path); {
	case err != nil:
		l.Z{l.E: err, "name": path}.Critical()
		return false
	default:
		return stat.Mode().Type() == fs.ModeSymlink
	}
}

func MustSymlink(oldname string, newname string) {
	switch err := os.Symlink(oldname, newname); {
	case errors.Is(err, fs.ErrExist):
		var (
			interim *os.LinkError
			_       = errors.As(err, &interim)
		)
		switch MustIsSymlink(newname) && interim.Old == oldname && interim.New == newname {
		case true:
			l.Z{l.E: err, "oldname": oldname, "newname": newname}.Debug()
			return
		}
	case err != nil:
		l.Z{l.E: err, "oldname": oldname, "newname": newname}.Critical()
	}
}

func MustRel(basepath string, targpath string) string {
	switch outbound, err := filepath.Rel(basepath, targpath); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}

func MustRemove(name string) {
	switch err := os.Remove(name); {
	case err != nil:
		l.Z{l.E: err}.Critical()
	}
}

func MustGetwd() string {
	switch outbound, err := os.Getwd(); {
	case err != nil:
		l.Z{l.E: err}.Critical()
		return ""
	default:
		return outbound
	}
}
