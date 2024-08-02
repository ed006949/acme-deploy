package l

import (
	"github.com/rs/zerolog"
)

type severity int

// type facility int

type F map[string]interface{}

type name string
type config string
type dryRun string
type verbosity string

type pControlStruct struct {
	name      string
	config    string
	dryRun    bool
	verbosity zerolog.Level
}
