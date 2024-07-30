package io_crypto

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"strings"

	"acme-deploy/src/l"
)

//
// taken from https://github.com/golang/go/src/crypto/tls/tls.go
// modified to be more useful
//

type SignatureScheme uint16

type Certificate struct {
	PrivateKeyDER   []byte
	CertificatesDER [][]byte
	// CertificateCAChainPEM    []byte

	PrivateKeyPEM         []byte
	CertificatesPEM       [][]byte
	CertificateCAChainDER []byte

	PrivateKeyRawPEM         []byte
	CertificatesRawPEM       [][]byte
	CertificateCAChainRawPEM []byte

	// Certificates is the parsed form of the leaf certificate, which may be initialized
	// using x509.ParseCertificate to reduce per-handshake processing. If nil,
	// the leaf certificate will be parsed as needed.
	Certificates []*x509.Certificate
	// PrivateKey contains the private key corresponding to the public key in
	// Certificates. This must implement crypto.Signer with an RSA, ECDSA or Ed25519 PublicKey.
	// For a server up to TLS 1.2, it can also implement crypto.Decrypter with
	// an RSA PublicKey.
	PrivateKey crypto.PrivateKey

	// SupportedSignatureAlgorithms is an optional list restricting what
	// signature algorithms the PrivateKey can be used for.
	SupportedSignatureAlgorithms []SignatureScheme
	// OCSPStaple contains an optional OCSP response which will be served
	// to clients that request it.
	OCSPStaple []byte
	// SignedCertificateTimestamps contains an optional list of Signed
	// Certificate Timestamps which will be served to clients that request it.
	SignedCertificateTimestamps [][]byte
}

func MustX509KeyPair(certPEMBlock, keyPEMBlock []byte) *Certificate {
	switch outbound, err := X509KeyPair(certPEMBlock, keyPEMBlock); {
	case err != nil:
		l.Error.E(err, nil)
		return nil
	default:
		return outbound
	}
}

func X509KeyPair(certPEMBlock []byte, keyPEMBlock []byte) (outbound *Certificate, err error) {
	var (
		certDERBlock *pem.Block
		keyDERBlock  *pem.Block
		interimCert  *x509.Certificate
	)
	outbound = new(Certificate)

	func() {
		for {
			certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
			// for ; certDERBlock != nil; certDERBlock, certPEMBlock = pem.Decode(certPEMBlock) {
			switch {
			case certDERBlock == nil:
				return
			case certDERBlock.Type == "CERTIFICATE":
				outbound.CertificatesDER = append(
					outbound.CertificatesDER,
					certDERBlock.Bytes,
				)
				outbound.CertificatesRawPEM = append(
					outbound.CertificatesRawPEM,
					[]byte(base64.RawStdEncoding.EncodeToString(certDERBlock.Bytes)),
				)

				switch {
				case len(outbound.CertificatesDER) > 1:
					outbound.CertificateCAChainDER = append(outbound.CertificateCAChainDER, certDERBlock.Bytes...)
				}
			}
		}
	}()
	switch {
	case len(outbound.CertificatesDER) == 0:
		return nil, errors.New("tls: failed to find any \"CERTIFICATE\" PEM block")
	}
	outbound.CertificateCAChainRawPEM = []byte(base64.RawStdEncoding.EncodeToString(outbound.CertificateCAChainDER))

	func() {
		for {
			keyDERBlock, keyPEMBlock = pem.Decode(keyPEMBlock)
			// for ; keyDERBlock != nil; keyDERBlock, keyPEMBlock = pem.Decode(keyPEMBlock) {
			switch {
			case keyDERBlock == nil:
				return
			case keyDERBlock.Type == "PRIVATE KEY" || strings.HasSuffix(keyDERBlock.Type, " PRIVATE KEY"):
				outbound.PrivateKeyDER = keyDERBlock.Bytes
				outbound.PrivateKeyRawPEM = []byte(base64.RawStdEncoding.EncodeToString(outbound.PrivateKeyDER))
			}
		}
	}()
	switch {
	case len(outbound.PrivateKeyDER) == 0:
		return nil, errors.New("tls: failed to find any \"PRIVATE KEY\" PEM block")
	}

	for _, b := range outbound.CertificatesDER {
		switch interimCert, err = x509.ParseCertificate(b); {
		case err != nil:
			return nil, err
		default:
			outbound.Certificates = append(outbound.Certificates, interimCert)
		}
	}

	switch outbound.PrivateKey, err = parsePrivateKey(outbound.PrivateKeyDER); {
	case err != nil:
		return nil, err
	}

	// TODO complete local chain verification
	// We don't need to parse the public key for TLS, but we so do anyway
	// to check that it looks sane and matches the private key.

	switch pub := outbound.Certificates[0].PublicKey.(type) {
	case *rsa.PublicKey:
		switch priv, ok := outbound.PrivateKey.(*rsa.PrivateKey); {
		case !ok:
			return nil, errors.New("tls: private key type does not match public key type")
		case pub.N.Cmp(priv.N) != 0:
			return nil, errors.New("tls: private key does not match public key")
		}
	case *ecdsa.PublicKey:
		switch priv, ok := outbound.PrivateKey.(*ecdsa.PrivateKey); {
		case !ok:
			return nil, errors.New("tls: private key type does not match public key type")
		case pub.X.Cmp(priv.X) != 0 || pub.Y.Cmp(priv.Y) != 0:
			return nil, errors.New("tls: private key does not match public key")
		}
	case ed25519.PublicKey:
		switch priv, ok := outbound.PrivateKey.(ed25519.PrivateKey); {
		case !ok:
			return nil, errors.New("tls: private key type does not match public key type")
		case !bytes.Equal(priv.Public().(ed25519.PublicKey), pub):
			return nil, errors.New("tls: private key does not match public key")
		}
	default:
		return nil, errors.New("tls: unknown public key algorithm")
	}

	return
}
func parsePrivateKey(der []byte) (key crypto.PrivateKey, err error) {
	switch key, err = x509.ParsePKCS1PrivateKey(der); {
	case err == nil:
		return
	}

	switch key, err = x509.ParsePKCS8PrivateKey(der); {
	case err == nil:
		switch value := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return value, nil
		default:
			return nil, errors.New("tls: found unknown private key type in PKCS#8 wrapping")
		}
	}

	switch key, err = x509.ParseECPrivateKey(der); {
	case err == nil:
		return
	}

	return nil, errors.New("tls: failed to parse private key")
}
