package l

import (
	"github.com/rs/zerolog"
)

type Z map[string]interface{}

type name string
type config string
type dryRun string
type verbosity string
type control string
type severity string
type facility string

type controlNumber int
type severityNumber int
type facilityNumber int

type controlStruct struct {
	name      string
	config    string
	dryRun    bool
	verbosity zerolog.Level
}
type packageStruct struct {
	name      string
	config    string
	dryRun    bool
	verbosity zerolog.Level
}
type logStruct struct {
	severity severityNumber
	facility facilityNumber
}

type nameType string
type configType string
type dryRunType bool
type verbosityType zerolog.Level
type controlType controlNumber
type severityType severityNumber
type facilityType facilityNumber
