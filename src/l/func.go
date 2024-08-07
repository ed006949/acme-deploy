package l

import (
	"flag"
	"net/url"
	"strings"
)

// func Emergency(e Z)     { log.Fatal().EmbedObject(e).Send() }
// func Alert(e Z)         { log.Fatal().EmbedObject(e).Send() }
// func Critical(e Z)      { log.Fatal().EmbedObject(e).Send() }
// func Error(e Z)         { log.Error().EmbedObject(e).Send() }
// func Warning(e Z)       { log.Warn().EmbedObject(e).Send() }
// func Notice(e Z)        { log.Info().EmbedObject(e).Send() }
// func Informational(e Z) { log.Info().EmbedObject(e).Send() }
// func Debug(e Z)         { log.Debug().EmbedObject(e).Send() }
// func Trace(e Z)         { log.Trace().EmbedObject(e).Send() }
// func Panic(e Z)         { log.Panic().EmbedObject(e).Send() }
// func Quiet(e Z)         {}
// func Disabled(e Z)      {}

func envDescription(inbound any) (outbound string) {
	switch value := inbound.(type) {
	case nameType:
		return "daemon name (" + value.EnvName() + ")"
	case configType:
		return "config file (" + value.EnvName() + ")"
	case dryRunType:
		return "dry-run flag, overrides config (" + value.EnvName() + ")"
	case verbosityType:
		return "verbosity level, overrides config (" + value.EnvName() + ")"
	case modeType:
		return "operational mode (" + value.EnvName() + ")"
	default:
		return
	}
}

func EnvName(inbound string) (outbound string) {
	return strings.ReplaceAll(
		strings.ToUpper(
			Name.String()+"_"+inbound),
		"-",
		"_",
	)
}

func ParseBool(inbound string) (bool, error) {
	switch {
	case len(inbound) == 0:
		return false, ENODATA
	}
	inbound = strings.ToLower(inbound)

	switch inbound {
	case "1", "t", "y", "true", "yes", "on":
		return true, nil
	case "0", "f", "n", "false", "no", "off":
		return false, nil
	default:
		return false, EINVAL
	}
}
func FormatBool(inbound bool) string {
	switch inbound {
	case true:
		return "true"
	default:
		return "false"
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
func IndexSlice[S ~[]E, E comparable, M map[E]int](inbound S) (outbound M) {
	outbound = make(M)
	for a, b := range inbound {
		outbound[b] = a
	}
	return
}

func StripErr(err error)                                 {}
func StripErr1[E any](inbound E, err error) (outbound E) { return inbound }

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
