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
func IndexSlice[S ~[]E, E comparable, M map[E]struct{}](inbound S) (outbound M) {
	outbound = make(M)
	for _, b := range inbound {
		outbound[b] = struct{}{}
	}
	return
}

func StripErr(err error) {
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
