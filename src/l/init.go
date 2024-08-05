package l

import (
	"github.com/rs/zerolog"
)

func init() {
	// l.log function call nesting depth is 1
	zerolog.CallerSkipFrameCount = zerolog.CallerSkipFrameCount + 1

	// parse defaults while init
	control.name.Set()
	control.config.Set()
	control.dryRun.Set()
	control.mode.Set()
	control.verbosity.Set()
}
