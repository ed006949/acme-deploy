package l

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
)

func init() {
	// l.log function call nesting depth is 1
	zerolog.CallerSkipFrameCount = zerolog.CallerSkipFrameCount + 1

	// parse defaults while init
	control.Name.Set()
	control.Config.Set()
	control.DryRun.Set()
	control.Mode.Set()
	control.Verbosity.Set()
}
func InitCLI() {
	var (
		cliName      = flag.String(Name.Name(), os.Getenv(Name.EnvName()), Name.EnvDescription())
		cliConfig    = flag.String(Config.Name(), os.Getenv(Config.EnvName()), Config.EnvDescription())
		cliDryRun    = flag.Bool(DryRun.Name(), StripErr1(ParseBool(os.Getenv(DryRun.EnvName()))), DryRun.EnvDescription())
		cliMode      = flag.String(Mode.Name(), os.Getenv(Mode.EnvName()), Mode.EnvDescription())
		cliVerbosity = flag.String(Verbosity.Name(), os.Getenv(Verbosity.EnvName()), Verbosity.EnvDescription())
	)
	flag.Parse()

	switch {
	case FlagIsFlagExist(Name.Name()):
		Name.Set(*cliName)
		fallthrough
	case FlagIsFlagExist(Config.Name()):
		Config.Set(*cliConfig)
		fallthrough
	case FlagIsFlagExist(DryRun.Name()):
		DryRun.Set(*cliDryRun)
		fallthrough
	case FlagIsFlagExist(Mode.Name()):
		Mode.Set(*cliMode)
		fallthrough
	case FlagIsFlagExist(Verbosity.Name()):
		Verbosity.Set(*cliVerbosity)
		fallthrough
	default:
	}
}
