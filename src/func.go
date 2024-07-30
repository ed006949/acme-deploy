package main

import (
	"errors"
	"strings"

	"github.com/go-ini/ini"

	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_vfs"
)

func load(vfsDB *io_vfs.VFSDB, iniFile string) (outbound *leConf, err error) {
	var (
		interimIniConf iniLEConf
		cert           *io_crypto.Certificate
	)

	switch err = ini.MapTo(&interimIniConf, vfsDB.MustReadFile(iniFile)); {
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
		cert:                cert,
	}
	// outbound.mxList = io_net.LookupMX(append(outbound.leAlt, interimIniConf.Le_Domain))

	// switch cert, err = io_crypto.X509KeyPair(
	// 	vfsDB.MustReadFile(outbound.leRealFullChainPath),
	// 	vfsDB.MustReadFile(outbound.leRealKeyPath),
	// ); {
	// case err != nil:
	// 	return nil, err
	// }

	//
	switch cert, err = io_crypto.X509KeyPair(
		vfsDB.MustReadFile(strings.ReplaceAll(outbound.leRealFullChainPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")),
		vfsDB.MustReadFile(strings.ReplaceAll(outbound.leRealKeyPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")),
	); {
	case err != nil:
		return nil, err
	}
	outbound.leRealCertPath = strings.ReplaceAll(outbound.leRealCertPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")
	outbound.leRealCACertPath = strings.ReplaceAll(outbound.leRealCACertPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")
	outbound.leRealKeyPath = strings.ReplaceAll(outbound.leRealKeyPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")
	outbound.leRealFullChainPath = strings.ReplaceAll(outbound.leRealFullChainPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/")
	//

	outbound.cert = cert

	return
}
