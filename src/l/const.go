package l

const (
	Name      name      = "name"
	Config    config    = "config"
	DryRun    dryRun    = "dry-run"
	Verbosity verbosity = "verbosity"
	Severity  severity  = "severity"
	Facility  facility  = "facility"
)

const (
	_emergency     severityNumber = iota // rfc3164
	_alert                               // rfc3164
	_critical                            // rfc3164
	_error                               // rfc3164
	_warning                             // rfc3164
	_notice                              // rfc3164
	_informational                       // rfc3164
	_debug                               // rfc3164
	_trace
	_panic
)

const (
	_kern         facilityNumber = iota // rfc3164
	_user                               // rfc3164
	_mail                               // rfc3164
	_daemon                             // rfc3164
	_auth                               // rfc3164
	_syslog                             // rfc3164
	_lpr                                // rfc3164
	_news                               // rfc3164
	_uucp                               // rfc3164
	_cron                               // rfc3164
	_authpriv                           // rfc3164
	_ftp                                // rfc3164
	_ntp                                // rfc3164
	_security                           // rfc3164
	_console                            // rfc3164
	_solaris_cron                       // rfc3164
	_local0                             // rfc3164
	_local1                             // rfc3164
	_local2                             // rfc3164
	_local3                             // rfc3164
	_local4                             // rfc3164
	_local5                             // rfc3164
	_local6                             // rfc3164
	_local7                             // rfc3164
)
