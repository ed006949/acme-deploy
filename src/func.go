package main

import (
	"bytes"
	"errors"

	"github.com/go-ini/ini"

	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_vfs"
)

func load(vfsDB *io_vfs.VFSDB, iniFile string) (outbound *leConf, err error) {
	var (
		interimIniConf iniLEConf
		cert           *io_crypto.Certificate
	)

	// switch err = ini.MapTo(&interimIniConf, vfsDB.MustReadFile(iniFile)); {
	switch err = ini.MapTo(&interimIniConf, bytes.ReplaceAll(
		vfsDB.MustReadFile(iniFile),
		[]byte("/var/etc/acme-client/"),
		[]byte(vfsDB.List["acme-client"]+"/"),
	)); {
	case err != nil:
		return nil, err
	case len(interimIniConf.Le_Domain) == 0 || len(interimIniConf.Le_RealCertPath) == 0 || len(interimIniConf.Le_RealCACertPath) == 0 || len(interimIniConf.Le_RealKeyPath) == 0 || len(interimIniConf.Le_RealFullChainPath) == 0:
		return nil, errors.New("config data is not enough")
	}

	outbound = &leConf{
		leDomain: interimIniConf.Le_Domain,
		leAlt: func() (outbound []string) {
			for _, f := range interimIniConf.Le_Alt {
				switch {
				case f == "no": // OPNSense and acme.sh, alt domain name = "no" ????
					continue
				}
				outbound = append(outbound, f)
			}
			return
		}(),
		leRealCertPath:      interimIniConf.Le_RealCertPath,
		leRealCACertPath:    interimIniConf.Le_RealCACertPath,
		leRealKeyPath:       interimIniConf.Le_RealKeyPath,
		leRealFullChainPath: interimIniConf.Le_RealFullChainPath,
		cert:                nil,
		mxList:              nil,
	}

	switch cert, err = io_crypto.X509KeyPair(
		vfsDB.MustReadFile(outbound.leRealFullChainPath),
		vfsDB.MustReadFile(outbound.leRealKeyPath),
	); {
	case err != nil:
		return nil, err
	}

	outbound.cert = cert
	// outbound.mxList = io_net.LookupMX(append(outbound.leAlt, interimIniConf.Le_Domain))

	return
}
