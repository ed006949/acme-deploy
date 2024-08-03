package l

import (
	"github.com/rs/zerolog"
)

var (
	pControl = &pStruct{
		name:      "",
		config:    "",
		dryRun:    false,
		verbosity: zerolog.InfoLevel,
		mode:      "",
	}
)
