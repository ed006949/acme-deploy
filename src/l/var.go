package l

import (
	"github.com/rs/zerolog"
)

var (
	pControl = &pStruct{
		name:      "",
		config:    "",
		verbosity: zerolog.InfoLevel,
		dryRun:    false,
	}
)
