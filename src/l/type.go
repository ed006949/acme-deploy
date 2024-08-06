package l

import (
	"github.com/rs/zerolog"
)

type Z map[string]interface{}

type nameType string
type configType string
type dryRunType string
type modeType string
type verbosityType string

type nameValue string
type configValue string
type dryRunFlag bool
type modeValue int
type verbosityLevel zerolog.Level

type ControlType struct {
	Name      nameValue      `xml:"name,attr,omitempty"`
	Config    configValue    `xml:"config,attr,omitempty"`
	DryRun    dryRunFlag     `xml:"dry-run,attr,omitempty"`
	Mode      modeValue      `xml:"mode,attr,omitempty"`
	Verbosity verbosityLevel `xml:"verbosity,attr,omitempty"`
}
