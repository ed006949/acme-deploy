package io_crypto

const (
	EX509ParsePrivKey errorNumber = iota
	EPEMNoDataKey
	EPEMNoDataCert
	EUnknownPrivKeyType
	EUnknownPubKeyAlgo
	EPrivKeyType
	EPrivKeySize
	ETypeMismatchPrivKeyPubKey
	EMismatchPrivKeyPubKey
)

var errorDescription = [...]string{
	EX509ParsePrivKey:          "x509: failed to parse private key",
	EPEMNoDataKey:              "PEM: failed to find any PRIVATE KEY data",
	EPEMNoDataCert:             "PEM: failed to find any CERTIFICATE data",
	EUnknownPrivKeyType:        "unknown private key type",
	EUnknownPubKeyAlgo:         "unknown public key algorithm",
	EPrivKeyType:               "wrong private key type",
	EPrivKeySize:               "wrong private key size",
	ETypeMismatchPrivKeyPubKey: "private key type does not match public key type",
	EMismatchPrivKeyPubKey:     "private key does not match public key",
}

type errorNumber int

func (e errorNumber) Error() string        { return errorDescription[e] }
func (e errorNumber) Is(target error) bool { return e == target }
