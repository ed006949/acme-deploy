package l

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (receiver severity) L(inbound map[string]interface{}) {
	receiver.logAction(nil, inbound)
}
func (receiver severity) E(err error, inbound map[string]interface{}) {
	receiver.logAction(err, inbound)
}

func (receiver severity) logAction(err error, inbound map[string]interface{}) {
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
		log.Error().Caller().Any("dryRun", PackageDryRun).Any("Severity", receiver).Err(EINVAL).Send()
		event = log.Error()
	}

	switch {
	case PackageDryRun:
		event.Any("dryRun", PackageDryRun)
	}
	event.AnErr("error", err).Fields(inbound).Send()
}
