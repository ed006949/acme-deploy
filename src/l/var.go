package l

var (
	control = &controlStruct{
		name:      "",
		config:    "",
		dryRun:    NoDryRun,
		verbosity: Informational,
		mode:      Init,
	}
)
var (
	dryRunDescription = map[dryRunFlag]string{
		NoDryRun: "false",
		DoDryRun: "true",
	}
)
var (
	modeDescription = map[modeValue]string{
		Init:   "init",
		Deploy: "deploy",
		CLI:    "cli",
		Daemon: "daemon",
	}
)

// var (
// 	verbosityDescription = map[verbosityLevel]string{
// 		// Emergency:     zerolog.Level(Emergency).String(),     // zerolog.FatalLevel
// 		// Alert:         zerolog.Level(Alert).String(),         // zerolog.FatalLevel
// 		Critical: zerolog.Level(Critical).String(), // zerolog.FatalLevel
// 		Error:    zerolog.Level(Error).String(),    // zerolog.ErrorLevel
// 		Warning:  zerolog.Level(Warning).String(),  // zerolog.WarnLevel
// 		// Notice:        zerolog.Level(Notice).String(),        // zerolog.InfoLevel
// 		Informational: zerolog.Level(Informational).String(), // zerolog.InfoLevel
// 		Debug:         zerolog.Level(Debug).String(),         // zerolog.DebugLevel
// 		Trace:         zerolog.Level(Trace).String(),         // zerolog.TraceLevel
// 		Panic:         zerolog.Level(Panic).String(),         // zerolog.PanicLevel
// 		Quiet:         zerolog.Level(Quiet).String(),         // zerolog.NoLevel
// 		Disabled:      zerolog.Level(Disabled).String(),      // zerolog.Disabled
// 	}
// )
