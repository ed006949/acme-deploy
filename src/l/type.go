package l

import (
	"github.com/rs/zerolog"
)

type Z map[string]interface{}

type nameType string
type configType string
type dryRunType string
type verbosityType string
type modeType string

type nameValue string
type configValue string
type dryRunFlag bool
type modeValue int
type verbosityLevel zerolog.Level

type controlStruct struct {
	name      nameValue
	config    configValue
	dryRun    dryRunFlag
	mode      modeValue
	verbosity verbosityLevel
}
