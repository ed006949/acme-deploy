package l

import (
	"github.com/rs/zerolog"
)

var (
	pControl = &pControlStruct{
		name:      "",
		config:    "",
		verbosity: zerolog.InfoLevel,
		dryRun:    false,
	}
)
