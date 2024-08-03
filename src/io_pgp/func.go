package io_pgp

import (
	"bytes"
	"io"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"

	"acme-deploy/src/l"
)

func mustArmorDecode(in io.Reader) *armor.Block {
	switch outbound, err := armor.Decode(in); {
	case err != nil:
		l.Critical(l.Z{"": err})
		return nil
	default:
		return outbound
	}
}
func mustReadEntity(block *armor.Block) *openpgp.Entity {
	switch outbound, err := openpgp.ReadEntity(packet.NewReader(block.Body)); {
	case err != nil:
		l.Critical(l.Z{"": err})
		return nil
	default:
		return outbound
	}
}
func mustDecryptPrivateKeys(e *openpgp.Entity, passphrase []byte) {
	switch err := e.DecryptPrivateKeys(passphrase); {
	case err != nil:
		l.Critical(l.Z{"": err})
	}
}
func mustGetEntity(data []byte, passphrase []byte) *openpgp.Entity {
	var (
		outbound = mustReadEntity(mustArmorDecode(bytes.NewReader(data)))
	)
	mustDecryptPrivateKeys(outbound, passphrase)
	return outbound
}
