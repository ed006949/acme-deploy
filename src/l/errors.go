package l

const (
	EORPHANED Errno = iota
	EDUPDATA
	EEXIST
	ENOTFOUND
	EINVAL
)

var errorDescription = [...]string{
	EORPHANED: "orphaned entry",
	EDUPDATA:  "duplicate data",
	EEXIST:    "already exists",
	ENOTFOUND: "not found",
	EINVAL:    "invalid argument",
}

type Errno uint

func (e Errno) Is(target error) bool { return e == target }
func (e Errno) Timeout() bool        { return false }
func (e Errno) Temporary() bool      { return e.Timeout() }
func (e Errno) Error() string        { return errorDescription[e] }

//
//
// custom error example
//
// type logError struct {
// 	Err      error
// 	Severity severity
// }
// func (e *logError) Error() string {
// 	return string(e.Severity) + "(" + strconv.Itoa(int(e.Severity)) + "), " + e.Err.Error()
// }
// func (e *logError) Unwrap() error { return e.Err }
// func newLogError(err error, severity severity) *logError {
// 	return &logError{
// 		Err:      err,
// 		Severity: severity,
// 	}
// }
//
