package l

var (
	control = &ControlType{
		Name:      "",
		Config:    "",
		DryRun:    DoDryRun,
		Mode:      Init,
		Verbosity: Informational,
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
