package l

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (receiver severity) L(inbound map[string]interface{}) {
	receiver.lAction(nil, inbound)
}
func (receiver severity) E(err error, inbound map[string]interface{}) {
	receiver.lAction(err, inbound)
}

func (receiver severity) lAction(err error, inbound map[string]interface{}) {
	var (
		event *zerolog.Event
	)

	switch receiver {
	case Panic:
		event = log.Panic()
	case Emergency:
		event = log.Fatal()
	case Alert:
		event = log.Fatal()
	case Critical:
		event = log.Fatal()
	case Error:
		event = log.Error()
	case Warning:
		event = log.Warn()
	case Notice:
		event = log.Info()
	case Informational:
		event = log.Info()
	case Debug:
		event = log.Debug()
	case Trace:
		event = log.Trace()
	default:
		log.Error().Caller().Any("dry-run", PackageDryRun).Any("Severity", receiver).Err(EINVAL).Send()
		event = log.Error()
	}

	switch {
	case PackageDryRun:
		event.Any("dry-run", PackageDryRun)
	}

	event.AnErr("error", err).Fields(inbound).Send()
}
