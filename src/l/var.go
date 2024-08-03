package l

import (
	"github.com/rs/zerolog"
)

var (
	pControl = &packageStruct{
		name:      "",
		config:    "",
		verbosity: zerolog.InfoLevel,
		dryRun:    false,
	}
)

var (
	lControl = &logStruct{
		facility: _daemon,
		severity: _informational,
	}
)

var (
	controlDescription = [...]string{
		_name:      "name",
		_config:    "config",
		_dryRun:    "dry-run",
		_verbosity: "verbosity",
	}
)
var (
	severityDescription = [...]string{
		_emergency:     "emergency",
		_alert:         "alert",
		_critical:      "critical",
		_error:         "error",
		_warning:       "warning",
		_notice:        "notice",
		_informational: "informational",
		_debug:         "debug",
		_trace:         "trace",
		_panic:         "panic",
	}
)

var (
	facilityDescription = [...]string{
		_kern:         "kern",
		_user:         "user",
		_mail:         "mail",
		_daemon:       "daemon",
		_auth:         "auth",
		_syslog:       "syslog",
		_lpr:          "lpr",
		_news:         "news",
		_uucp:         "uucp",
		_cron:         "cron",
		_authpriv:     "authpriv",
		_ftp:          "ftp",
		_ntp:          "ntp",
		_security:     "security",
		_console:      "console",
		_solaris_cron: "solaris_cron",
		_local0:       "local0",
		_local1:       "local1",
		_local2:       "local2",
		_local3:       "local3",
		_local4:       "local4",
		_local5:       "local5",
		_local6:       "local6",
		_local7:       "local7",
	}
)
