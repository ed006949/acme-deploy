package io_cgp

const (
	ECom Errno = iota
	EComSet
	EComSetDomAdm
	EComSetDomSetAdm
)

var errorDescription = [...]string{
	ECom:             "unknown command",
	EComSet:          "unknown command set",
	EComSetDomAdm:    "unknown Domain Administration command",
	EComSetDomSetAdm: "unknown Domain Set Administration command",
}

type Errno uint

func (e Errno) Is(target error) bool { return e == target }
func (e Errno) Timeout() bool        { return false }
func (e Errno) Temporary() bool      { return e.Timeout() }
func (e Errno) Error() string        { return errorDescription[e] }
