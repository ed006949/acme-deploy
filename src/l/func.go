package l

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func SetPackageVerbosity(inbound string) error {
	switch value, err := zerolog.ParseLevel(inbound); {
	case err != nil:
		return err
	case len(inbound) == 0 || value == zerolog.NoLevel:
		return syscall.EINVAL
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
		return errors.New("unknown dry-run variable type")
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
		return false, errors.New("unknown bool string")
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
