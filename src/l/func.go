package l

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetPackageVerbosity(inbound string) error {
	switch value, err := zerolog.ParseLevel(inbound); {
	case err != nil:
		return err
	case len(inbound) == 0 || value == zerolog.NoLevel:
		return EINVAL
	default:
		zerolog.SetGlobalLevel(value) // how it works ....
		log.Logger = log.Level(value).With().Timestamp().Caller().Logger().Output(zerolog.ConsoleWriter{
			Out:              os.Stderr,
			NoColor:          false,
			TimeFormat:       time.RFC3339,
			FormatFieldValue: func(i interface{}) string { return fmt.Sprintf("\"%s\"", i) },
		})
		return nil
	}
}

func SetPackageDryRun(inbound any) error {
	switch inboundValue := inbound.(type) {
	case string:
		switch value, err := ParseBool(inboundValue); {
		case err != nil:
			return err
		default:
			PackageDryRun = value
			return nil
		}
	case bool:
		PackageDryRun = inboundValue
		return nil
	default:
		return EINVAL
	}
}

func SetPackageName(inbound string) error {
	PackageName = inbound
	return nil
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

func IsFlagExist(name string) (outbound bool) {
	flag.Visit(func(fn *flag.Flag) {
		switch {
		case fn.Name == name:
			outbound = true
		}
	})
	return
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

func StripErr1[E comparable](inbound E, err error) (outbound E) {
	return inbound
}
