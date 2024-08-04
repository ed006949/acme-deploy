package l

import (
	"github.com/rs/zerolog"
)

func init() {
	// l.log function call nesting depth is 1
	zerolog.CallerSkipFrameCount = zerolog.CallerSkipFrameCount + 1

	// parse defaults while init
	Name.Set(pControl.name)
	Config.Set(pControl.config)
	Verbosity.Set(pControl.verbosity)
	DryRun.Set(pControl.dryRun)
	Mode.Set(pControl.mode)
}
