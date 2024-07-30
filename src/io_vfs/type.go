package io_vfs

import (
	"github.com/avfs/avfs"
)

type VFSDB struct {
	List map[string]string
	VFS  avfs.VFS
}
