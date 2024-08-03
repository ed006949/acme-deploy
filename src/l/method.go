package l

import (
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (r Z) Emergency()     { log.Fatal().EmbedObject(r).Send() }
func (r Z) Alert()         { log.Fatal().EmbedObject(r).Send() }
func (r Z) Critical()      { log.Fatal().EmbedObject(r).Send() }
func (r Z) Error()         { log.Error().EmbedObject(r).Send() }
func (r Z) Warning()       { log.Warn().EmbedObject(r).Send() }
func (r Z) Notice()        { log.Info().EmbedObject(r).Send() }
func (r Z) Informational() { log.Info().EmbedObject(r).Send() }
func (r Z) Debug()         { log.Debug().EmbedObject(r).Send() }
func (r Z) Trace()         { log.Trace().EmbedObject(r).Send() }
func (r Z) Panic()         { log.Panic().EmbedObject(r).Send() }

func (r Z) MarshalZerologObject(e *zerolog.Event) {
	for a, b := range r {
		switch a {
		case E:
			a = zerolog.ErrorFieldName
		case M:
			a = zerolog.MessageFieldName
		}

		switch value := b.(type) {
		case name:
			e.Str(value.Name(), value.Value())
		case config:
			e.Str(value.Name(), value.Value())
		case dryRun:
			e.Bool(value.Name(), value.Value())
		case verbosity:
			e.Str(value.Name(), value.String())
		case severity:
			e.Str(value.Name(), value.String())
		case facility:
			e.Str(value.Name(), value.String())
		case severityNumber:
			e.Str(a, value.String())
		case facilityNumber:
			e.Str(a, value.String())
		case error:
			e.AnErr(a, value)
		default:
			e.Interface(a, value)
		}
	}
	switch {
	case DryRun.Value():
		e.Bool(DryRun.Name(), DryRun.Value())
	}
}

func (r name) Set(inbound string)             { pControl.name = inbound }
func (r config) Set(inbound string)           { pControl.config = inbound }
func (r dryRun) Set(inbound bool)             { pControl.dryRun = inbound }
func (r verbosity) Set(inbound zerolog.Level) { setVerbosity(inbound) }
func (r severity) Set(inbound severityNumber) { lControl.severity = inbound }
func (r facility) Set(inbound facilityNumber) { lControl.facility = inbound }

func (r dryRun) SetString(inbound string) error {
	switch value, err := ParseBool(inbound); {
	case err != nil:
		return err
	default:
		pControl.dryRun = value
		return nil
	}
}
func (r verbosity) SetString(inbound string) error {
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

func (r name) Value() string             { return pControl.name }      // Package Flag Value
func (r config) Value() string           { return pControl.config }    // Package Flag Value
func (r dryRun) Value() bool             { return pControl.dryRun }    // Package Flag Value
func (r verbosity) Value() zerolog.Level { return pControl.verbosity } // Package Flag Value
func (r severity) Value() severityNumber { return lControl.severity }  // Package Flag Value
func (r facility) Value() facilityNumber { return lControl.facility }  // Package Flag Value

func (r name) String() string           { return pControl.name }                          // Package Flag String Value
func (r config) String() string         { return pControl.config }                        // Package Flag String Value
func (r dryRun) String() string         { return strconv.FormatBool(pControl.dryRun) }    // Package Flag String Value
func (r verbosity) String() string      { return pControl.verbosity.String() }            // Package Flag String Value
func (r severity) String() string       { return severityDescription[lControl.severity] } // Package Flag String Value
func (r facility) String() string       { return facilityDescription[lControl.facility] } // Package Flag String Value
func (r controlNumber) String() string  { return controlDescription[r] }                  // String Value
func (r severityNumber) String() string { return severityDescription[r] }                 // String Value
func (r facilityNumber) String() string { return facilityDescription[r] }                 // String Value

func (r name) Name() string      { return string(Name) }      // Package Flag Name
func (r config) Name() string    { return string(Config) }    // Package Flag Name
func (r dryRun) Name() string    { return string(DryRun) }    // Package Flag Name
func (r verbosity) Name() string { return string(Verbosity) } // Package Flag Name
func (r control) Name() string   { return string(Control) }   // Package Flag Name
func (r severity) Name() string  { return string(Severity) }  // Package Flag Name
func (r facility) Name() string  { return string(Facility) }  // Package Flag Name
