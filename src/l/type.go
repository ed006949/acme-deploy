package l

import (
	"github.com/rs/zerolog"
)

type Z map[string]interface{}

type name string
type config string
type dryRun string
type verbosity string
type mode string

type pStruct struct {
	name      string
	config    string
	dryRun    bool
	verbosity zerolog.Level
	mode      string
}
