package main

import (
	"bytes"
	"errors"

	"github.com/go-ini/ini"

	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/l"
)

func load(vfsDB *io_vfs.VFSDB, iniFile string) (outbound *leConf, err error) {
	outbound = new(leConf)

	// switch err = ini.MapTo(&interimIniConf, vfsDB.MustReadFile(iniFile)); {
	switch err = ini.MapTo(outbound, bytes.ReplaceAll(
		vfsDB.MustReadFile(iniFile),
		[]byte("/var/etc/acme-client/"),
		[]byte(vfsDB.List["acme-client"]+"/"),
	)); {
	case err != nil:
		return nil, err
	case len(outbound.LEDomain) == 0 || len(outbound.LERealCertPath) == 0 || len(outbound.LERealCACertPath) == 0 || len(outbound.LERealKeyPath) == 0 || len(outbound.LERealFullChainPath) == 0:
		return nil, errors.New("config data is not enough")
	}

	outbound.LEAlt = l.FilterSlice(outbound.LEAlt, "no") // OPNSense and acme.sh, alt domain name = "no" ????

	switch outbound.Certificate, err = io_crypto.X509KeyPair(
		vfsDB.MustReadFile(outbound.LERealFullChainPath),
		vfsDB.MustReadFile(outbound.LERealKeyPath),
	); {
	case err != nil:
		return nil, err
	}

	// outbound.MXList = io_net.LookupMX(append(outbound.LEAlt, interimIniConf.LEDomain))

	return
}
