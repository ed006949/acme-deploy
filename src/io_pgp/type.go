package io_pgp

import (
	"github.com/ProtonMail/go-crypto/openpgp"
)

type SignDB map[string]*openpgp.Entity
