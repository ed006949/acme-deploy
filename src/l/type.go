package l

import (
	"github.com/rs/zerolog"
)

type Z map[string]interface{}

type name string
type config string
type dryRun string
type verbosity string
type severity string
type facility string
type severityNumber int
type facilityNumber int

type pControlStruct struct {
	name      string
	config    string
	dryRun    bool
	verbosity zerolog.Level
}
type lControlStruct struct {
	severity severityNumber
	facility facilityNumber
}
