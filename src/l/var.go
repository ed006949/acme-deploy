package l

var (
	control = &controlType{
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
