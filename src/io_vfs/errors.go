package io_vfs

import (
	"errors"
)

var (
	ErrOrphanedEntry  = errors.New("orphaned entry")
	ErrListIDNotFound = errors.New("list ID not found")
)
