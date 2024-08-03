package l

import (
	"strconv"

	"github.com/rs/zerolog"
)

func (receiver Z) MarshalZerologObject(e *zerolog.Event) {
	for a, b := range receiver {
		switch value := b.(type) {
		case name:
			e.Str(value.Name(), value.Value())
		case config:
			e.Str(value.Name(), value.Value())
		case dryRun:
			// e.Bool(value.Name(), value.Value())
		case verbosity:
			e.Str(value.Name(), value.String())
		case severity:
			e.Str(value.Name(), value.String())
		case facility:
			e.Str(value.Name(), value.String())
		case severityNumber:
			e.Str(Severity.Name(), severityDescription[value])
		case facilityNumber:
			e.Str(Facility.Name(), facilityDescription[value])
		case error:
			e.AnErr("error", value)
		case string:
			e.Str(a, value)
		case bool:
			e.Bool(a, value)
		default:
			e.Interface(a, value)
		}
	}
	switch {
	case DryRun.Value():
		e.Bool(DryRun.Name(), DryRun.Value())
	}
}

func (receiver name) Set(inbound string)             { pControl.name = inbound }
func (receiver config) Set(inbound string)           { pControl.config = inbound }
func (receiver dryRun) Set(inbound bool)             { pControl.dryRun = inbound }
func (receiver verbosity) Set(inbound zerolog.Level) { setVerbosity(inbound) }
func (receiver severity) Set(inbound severityNumber) { lControl.severity = inbound }
func (receiver facility) Set(inbound facilityNumber) { lControl.facility = inbound }

func (receiver dryRun) SetString(inbound string) error {
	switch value, err := ParseBool(inbound); {
	case err != nil:
		return err
	default:
		pControl.dryRun = value
		return nil
	}
}
func (receiver verbosity) SetString(inbound string) error {
	switch value, err := zerolog.ParseLevel(inbound); {
	case err != nil:
		return err
	case len(inbound) == 0 || value == zerolog.NoLevel:
		return EINVAL
	default:
		setVerbosity(value)
		return nil
	}
}

func (receiver name) Value() string             { return pControl.name }      // Package Flag Value
func (receiver config) Value() string           { return pControl.config }    // Package Flag Value
func (receiver dryRun) Value() bool             { return pControl.dryRun }    // Package Flag Value
func (receiver verbosity) Value() zerolog.Level { return pControl.verbosity } // Package Flag Value
func (receiver severity) Value() severityNumber { return lControl.severity }  // Package Flag Value
func (receiver facility) Value() facilityNumber { return lControl.facility }  // Package Flag Value

func (receiver name) String() string      { return pControl.name }                          // Package Flag String Value
func (receiver config) String() string    { return pControl.config }                        // Package Flag String Value
func (receiver dryRun) String() string    { return strconv.FormatBool(pControl.dryRun) }    // Package Flag String Value
func (receiver verbosity) String() string { return pControl.verbosity.String() }            // Package Flag String Value
func (receiver severity) String() string  { return severityDescription[lControl.severity] } // Package Flag String Value
func (receiver facility) String() string  { return facilityDescription[lControl.facility] } // Package Flag String Value

func (receiver name) Name() string      { return string(Name) }      // Package Flag Name
func (receiver config) Name() string    { return string(Config) }    // Package Flag Name
func (receiver dryRun) Name() string    { return string(DryRun) }    // Package Flag Name
func (receiver verbosity) Name() string { return string(Verbosity) } // Package Flag Name
func (receiver severity) Name() string  { return string(Severity) }  // Package Flag Name
func (receiver facility) Name() string  { return string(Facility) }  // Package Flag Name
