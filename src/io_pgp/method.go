package io_pgp

import (
	"github.com/ProtonMail/go-crypto/openpgp"

	"acme-deploy/src/l"
)

func (receiver *SignDB) MustWriteSign(name string, data []byte, passphrase []byte) {
	switch _, ok := (*receiver)[name]; {
	case ok:
		l.Critical.E(l.EDUPDATA, l.F{"sign": name})
		return
	}
	(*receiver)[name] = mustGetEntity(data, passphrase)
}
func (receiver *SignDB) MustReadSign(name string) *openpgp.Entity {
	switch value, ok := (*receiver)[name]; {
	case !ok:
		l.Critical.E(l.ENOTFOUND, l.F{"sign": name})
		return nil
	default:
		return value
	}
}
