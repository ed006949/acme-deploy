package io_pgp

import (
	"bytes"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"

	"acme-deploy/src/l"
)

func (r *SignDB) ReadSign(name string) (outbound *openpgp.Entity, err error) {
	switch value, ok := (*r)[name]; {
	case !ok:
		return nil, l.ENOTFOUND
	default:
		return value, nil
	}
}

func (r *SignDB) WriteSign(name string, data []byte, passphrase []byte) (err error) {
	switch _, ok := (*r)[name]; {
	case ok:
		return l.EDUPDATA
	}

	var (
		armorBlock    *armor.Block
		openpgpEntity *openpgp.Entity
	)
	switch armorBlock, err = armor.Decode(bytes.NewReader(data)); {
	case err != nil:
		return
	}
	switch openpgpEntity, err = openpgp.ReadEntity(packet.NewReader(armorBlock.Body)); {
	case err != nil:
		return
	}
	switch err = openpgpEntity.DecryptPrivateKeys(passphrase); {
	case err != nil:
		return
	}

	(*r)[name] = openpgpEntity
	return
}
