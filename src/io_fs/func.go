package io_fs

import (
	"errors"
	"io/fs"
	"os"
)

func IsExist(inbound string) (outbound bool, err error) {
	switch _, err = os.Stat(inbound); {
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func Symlink(oldname string, newname string) (err error) {
	switch err = os.Symlink(oldname, newname); {
	case errors.Is(err, fs.ErrExist):
		var (
			interim    *os.LinkError
			_          = errors.As(err, &interim)
			isLink     bool
			fsFileinfo fs.FileInfo
		)
		switch fsFileinfo, err = os.Lstat(newname); {
		case err != nil:
			return
		}

		switch isLink = fsFileinfo.Mode().Type() == fs.ModeSymlink; {
		case isLink && interim.Old == oldname && interim.New == newname: // symlink exists and matches
			return nil
		}

		switch err = os.Remove(newname); {
		case errors.Is(err, fs.ErrNotExist): // why is that?
		case err != nil:
			return
		}

		switch err = os.Symlink(oldname, newname); {
		case err != nil:
			return
		}

		return
	case err != nil:
		return

	default:
		return
	}
}
