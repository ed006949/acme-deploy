package l

// custom error example

// import (
// 	"strconv"
// )
//
// type logError struct {
// 	Err      error
// 	Severity severity
// }
//
// func (e *logError) Error() string {
// 	return e.Severity.String() + "(" + strconv.Itoa(int(e.Severity)) + "), " + e.Err.Error()
// }
// func (e *logError) Unwrap() error { return e.Err }
//
// func newLogError(err error, severity severity) *logError {
// 	return &logError{
// 		Err:      err,
// 		Severity: severity,
// 	}
// }
