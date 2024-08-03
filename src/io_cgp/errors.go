package io_cgp

const (
	ECom errorNumber = iota
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

type errorNumber int

func (e errorNumber) Error() string        { return errorDescription[e] }
func (e errorNumber) Is(target error) bool { return e == target }
