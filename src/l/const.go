package l

const (
	Panic         severity = iota - 1
	Emergency              // rfc3164
	Alert                  // rfc3164
	Critical               // rfc3164
	Error                  // rfc3164
	Warning                // rfc3164
	Notice                 // rfc3164
	Informational          // rfc3164
	Debug                  // rfc3164
	Trace
)

// const (
// 	Kern         facility = iota // rfc3164
// 	User                         // rfc3164
// 	Mail                         // rfc3164
// 	Daemon                       // rfc3164
// 	Auth                         // rfc3164
// 	Syslog                       // rfc3164
// 	LPR                          // rfc3164
// 	News                         // rfc3164
// 	UUCP                         // rfc3164
// 	Cron                         // rfc3164
// 	Authpriv                     // rfc3164
// 	FTP                          // rfc3164
// 	NTP                          // rfc3164
// 	Security                     // rfc3164
// 	Console                      // rfc3164
// 	Solaris_cron                 // rfc3164
// 	Local0                       // rfc3164
// 	Local1                       // rfc3164
// 	Local2                       // rfc3164
// 	Local3                       // rfc3164
// 	Local4                       // rfc3164
// 	Local5                       // rfc3164
// 	Local6                       // rfc3164
// 	Local7                       // rfc3164
// )
