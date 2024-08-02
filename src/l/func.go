package l

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (receiver name) Set(inbound string)             { pControl.name = inbound }
func (receiver config) Set(inbound string)           { pControl.config = inbound }
func (receiver dryRun) Set(inbound bool)             { pControl.dryRun = inbound }
func (receiver verbosity) Set(inbound zerolog.Level) { setVerbosity(inbound) }

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

func (receiver name) String() string      { return pControl.name }                       // Package Flag String Value
func (receiver config) String() string    { return pControl.config }                     // Package Flag String Value
func (receiver dryRun) String() string    { return strconv.FormatBool(pControl.dryRun) } // Package Flag String Value
func (receiver verbosity) String() string { return pControl.verbosity.String() }         // Package Flag String Value

func (receiver name) Name() string      { return string(Name) }      // Package Flag Name
func (receiver config) Name() string    { return string(Config) }    // Package Flag Name
func (receiver dryRun) Name() string    { return string(DryRun) }    // Package Flag Name
func (receiver verbosity) Name() string { return string(Verbosity) } // Package Flag Name

func setVerbosity(inbound zerolog.Level) {
	pControl.verbosity = inbound
	zerolog.SetGlobalLevel(pControl.verbosity) // how it works ....
	log.Logger = log.Level(pControl.verbosity).With().Timestamp().Caller().Logger().Output(zerolog.ConsoleWriter{
		Out:              os.Stderr,
		NoColor:          false,
		TimeFormat:       time.RFC3339,
		FormatFieldValue: func(i interface{}) string { return fmt.Sprintf("\"%s\"", i) },
	})
}

func ParseBool(inbound string) (bool, error) {
	switch strings.ToLower(inbound) {
	case "1", "t", "y", "true", "yes", "on":
		return true, nil
	case "0", "f", "n", "false", "no", "off":
		return false, nil
	default:
		return false, EINVAL
	}
}

func FilterSlice[S ~[]E, E comparable](inbound S, filter ...E) (outbound S) {
	var (
		interim = IndexSlice(filter)
	)
	for _, b := range inbound {
		switch _, ok := interim[b]; {
		case !ok:
			outbound = append(outbound, b)
		}
	}
	return
}
func IndexSlice[S ~[]E, E comparable, M map[E]struct{}](inbound S) (outbound M) {
	outbound = make(M)
	for _, b := range inbound {
		outbound[b] = struct{}{}
	}
	return
}

func StripErr0(err error) {
	// Debug.E(err, nil)
}
func StripErr1[E comparable](inbound E, err error) (outbound E) {
	// Debug.E(err, nil)
	return inbound
}

func FlagIsFlagExist(name string) (outbound bool) {
	flag.Visit(func(fn *flag.Flag) {
		switch {
		case fn.Name == name:
			outbound = true
		}
	})
	return
}

func UrlParse(inbound string) (outbound *url.URL, err error) {
	switch outbound, err = url.Parse(inbound); {
	case err != nil:
		return nil, err
	case len(outbound.String()) == 0:
		return nil, ENODATA
	default:
		return outbound, nil
	}
}
