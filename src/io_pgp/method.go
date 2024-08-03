package io_pgp

import (
	"github.com/ProtonMail/go-crypto/openpgp"

	"acme-deploy/src/l"
)

func (r *SignDB) MustWriteSign(name string, data []byte, passphrase []byte) {
	switch _, ok := (*r)[name]; {
	case ok:
		l.Z{l.E: l.EDUPDATA, "sign": name}.Critical()
		return
	}
	(*r)[name] = mustGetEntity(data, passphrase)
}
func (r *SignDB) MustReadSign(name string) *openpgp.Entity {
	switch value, ok := (*r)[name]; {
	case !ok:
		l.Z{l.E: l.ENOTFOUND, "sign": name}.Critical()
		return nil
	default:
		return value
	}
}
