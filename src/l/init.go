package l

import (
	"github.com/rs/zerolog"
)

func init() {
	// log function call nesting depth is 2
	zerolog.CallerSkipFrameCount = zerolog.CallerSkipFrameCount + 2

	// set default log level while init
	_ = SetPackageVerbosity(zerolog.LevelInfoValue)
}
