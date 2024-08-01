package l

const (
	EORPHANED errorNumber = iota
	EDUPDATA
	EEXIST
	ENOTFOUND
	EINVAL
	ENEDATA
	ENOCONF
	EUEDATA
)

var errorDescription = [...]string{
	EORPHANED: "orphaned entry",
	EDUPDATA:  "duplicate data",
	EEXIST:    "already exists",
	ENOTFOUND: "not found",
	EINVAL:    "invalid argument",
	ENEDATA:   "not enough data",
	ENOCONF:   "not config",
	EUEDATA:   "unexpected data",
}

type errorNumber uint

func (e errorNumber) Error() string        { return errorDescription[e] }
func (e errorNumber) Is(target error) bool { return e == target }

// func (e errorNumber) Timeout() bool        { return false }
// func (e errorNumber) Temporary() bool      { return e.Timeout() }

// func (e errorNumber) Wrap() error          { return fmt.Errorf("%w", errors.New(e.Error())) }
// func (e errorNumber) Unwrap() error        { return e }

// fmt.Errorf("%w %w", errors.New("foo"), errors.New("bar"))

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
