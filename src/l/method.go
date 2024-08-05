package l

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (r Z) MarshalZerologObject(e *zerolog.Event) {
	for a, b := range r {
		// switch a {
		// case E:
		// 	a = zerolog.ErrorFieldName
		// case M:
		// 	a = zerolog.MessageFieldName
		// 	// case T:
		// 	// 	a = zerolog.TypeFieldName
		// }

		switch value := b.(type) {
		case nameValue:
			e.Str(a, value.String())
		case configValue:
			e.Str(a, value.String())
		case dryRunFlag:
			e.Bool(a, value.Flag())
		case modeValue:
			e.Str(a, value.String())
		case verbosityLevel:
			e.Str(a, value.String())
		case error:
			e.AnErr(a, value)
		default:
			switch a {
			case T:
				e.Type(a, b)
			default:
				e.Interface(a, value)
			}
		}
	}

	switch {
	case DryRun.Flag():
		e.Bool(DryRun.Name(), DryRun.Flag())
	}
}

func (r Z) Emergency()     { log.Fatal().EmbedObject(r).Send() } // rfc3164 ----
func (r Z) Alert()         { log.Fatal().EmbedObject(r).Send() } // rfc3164 ----
func (r Z) Critical()      { log.Fatal().EmbedObject(r).Send() } // rfc3164 ----
func (r Z) Error()         { log.Error().EmbedObject(r).Send() } // rfc3164 +
func (r Z) Warning()       { log.Warn().EmbedObject(r).Send() }  // rfc3164 +
func (r Z) Notice()        { log.Info().EmbedObject(r).Send() }  // rfc3164 ----
func (r Z) Informational() { log.Info().EmbedObject(r).Send() }  // rfc3164 +
func (r Z) Debug()         { log.Debug().EmbedObject(r).Send() } // rfc3164 +
func (r Z) Trace()         { log.Trace().EmbedObject(r).Send() } // specific +
func (r Z) Panic()         { log.Panic().EmbedObject(r).Send() } // specific +
func (r Z) Quiet()         {}                                    // specific +
func (r Z) Disabled()      {}                                    // specific +

func (r nameValue) Set()      { control.set(Name, r) }      // Package predefined Flag hook
func (r configValue) Set()    { control.set(Config, r) }    // Package predefined Flag hook
func (r dryRunFlag) Set()     { control.set(DryRun, r) }    // Package predefined Flag hook
func (r modeValue) Set()      { control.set(Mode, r) }      // Package predefined Flag hook
func (r verbosityLevel) Set() { control.set(Verbosity, r) } // Package predefined Flag hook

func (r nameType) Set(inbound any)      { control.set(Name, inbound) }      // Package Flag hook
func (r configType) Set(inbound any)    { control.set(Config, inbound) }    // Package Flag hook
func (r dryRunType) Set(inbound any)    { control.set(DryRun, inbound) }    // Package Flag hook
func (r modeType) Set(inbound any)      { control.set(Mode, inbound) }      // Package Flag hook
func (r verbosityType) Set(inbound any) { control.set(Verbosity, inbound) } // Package Flag hook

func (r nameType) Name() string      { return string(Name) }      // Package Flag Name
func (r configType) Name() string    { return string(Config) }    // Package Flag Name
func (r dryRunType) Name() string    { return string(DryRun) }    // Package Flag Name
func (r modeType) Name() string      { return string(Mode) }      // Package Flag Name
func (r verbosityType) Name() string { return string(Verbosity) } // Package Flag Name

func (r nameType) Value() nameValue           { return control.name }      // Package Flag Value
func (r configType) Value() configValue       { return control.config }    // Package Flag Value
func (r dryRunType) Value() dryRunFlag        { return control.dryRun }    // Package Flag Value
func (r modeType) Value() modeValue           { return control.mode }      // Package Flag Value
func (r verbosityType) Value() verbosityLevel { return control.verbosity } // Package Flag Value

func (r dryRunType) Flag() bool              { return r.Value().Flag() }   // Package Flag Flag Value
func (r verbosityType) Level() zerolog.Level { return r.Value().Level() }  // Package Flag Level Value
func (r nameType) String() string            { return r.Value().String() } // Package Flag String Value
func (r configType) String() string          { return r.Value().String() } // Package Flag String Value
func (r dryRunType) String() string          { return r.Value().String() } // Package Flag String Value
func (r modeType) String() string            { return r.Value().String() } // Package Flag String Value
func (r verbosityType) String() string       { return r.Value().String() } // Package Flag String Value

func (r dryRunFlag) Flag() bool               { return bool(r) }                   // Package Flag flag
func (r verbosityLevel) Level() zerolog.Level { return zerolog.Level(r) }          // Package Flag level
func (r nameValue) String() string            { return string(r) }                 // Package Flag description
func (r configValue) String() string          { return string(r) }                 // Package Flag description
func (r dryRunFlag) String() string           { return dryRunDescription[r] }      // Package Flag description
func (r modeValue) String() string            { return modeDescription[r] }        // Package Flag description
func (r verbosityLevel) String() string       { return zerolog.Level(r).String() } // Package Flag description

func (r controlStruct) set(inboundKey any, inboundValue any) {
	switch inboundKey.(type) {
	case nameType:
		switch value := inboundValue.(type) {
		case nameValue:
			control.name = value
		case string:
			control.name = nameValue(value)
		}

	case configType:
		switch value := inboundValue.(type) {
		case configValue:
			control.config = value
		case string:
			control.config = configValue(value)
		}

	case dryRunType:
		switch value := inboundValue.(type) {
		case dryRunFlag:
			control.dryRun = value
		case bool:
			control.dryRun = dryRunFlag(value)

		case string:
			switch {
			case len(value) == 0:
				return
			}
			value = strings.ToLower(value)
			switch value {
			case "1", "t", "y", "true", "yes", "on":
				control.dryRun = true
			case "0", "f", "n", "false", "no", "off":
				control.dryRun = false
			}
		}

	case modeType:
		switch value := inboundValue.(type) {
		case modeValue:
			control.mode = value
		case int:
			control.mode = modeValue(value)

		case string:
			switch {
			case len(value) == 0:
				return
			}
			value = strings.ToLower(value)
			for a, b := range modeDescription {
				switch {
				case value == b:
					control.mode = a
					return
				}
			}
		}

	case verbosityType:
		switch value := inboundValue.(type) {
		case verbosityLevel:
			control.verbosity = value
		case int8:
			control.verbosity = verbosityLevel(value)

		case zerolog.Level:
			control.verbosity = verbosityLevel(value)

		case string:
			switch {
			case len(value) == 0:
				return
			}
			value = strings.ToLower(value)
			switch interim, err := zerolog.ParseLevel(value); {
			case err != nil:
				return
			default:
				control.verbosity = verbosityLevel(interim)
			}
		}

		zerolog.SetGlobalLevel(control.verbosity.Level()) // .!. how it works ....
		log.Logger = log.Level(control.verbosity.Level()).With().Timestamp().Caller().Logger().Output(zerolog.ConsoleWriter{
			Out:              os.Stderr,
			NoColor:          false,
			TimeFormat:       time.RFC3339,
			FormatFieldValue: func(i interface{}) string { return fmt.Sprintf("\"%s\"", i) },
		})

	}
}
