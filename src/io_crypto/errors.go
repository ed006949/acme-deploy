package io_crypto

const (
	EX509ParsePrivKey errorNumber = iota
	EPEMNoDataKey
	EPEMNoDataCert
	EUnknownTypePrivKey
	EUnknownAlgoPubKey
	ETypeMismatchPrivKeyPubKey
	EMismatchPrivKeyPubKey
)

var errorDescription = [...]string{
	EX509ParsePrivKey:          "x509: failed to parse private key",
	EPEMNoDataKey:              "PEM: failed to find any PRIVATE KEY data",
	EPEMNoDataCert:             "PEM: failed to find any CERTIFICATE data",
	EUnknownTypePrivKey:        "unknown private key type",
	EUnknownAlgoPubKey:         "unknown public key algorithm",
	ETypeMismatchPrivKeyPubKey: "private key type does not match public key type",
	EMismatchPrivKeyPubKey:     "private key does not match public key",
}

type errorNumber uint

func (e errorNumber) Error() string        { return errorDescription[e] }
func (e errorNumber) Is(target error) bool { return e == target }
