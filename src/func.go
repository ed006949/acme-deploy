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

	switch cert, err = io_crypto.X509KeyPair(
		// 	vfsDB.MustReadFile(interimIniConf.Le_RealFullChainPath),
		// 	vfsDB.MustReadFile(interimIniConf.Le_RealKeyPath),
		vfsDB.MustReadFile(strings.ReplaceAll(
			interimIniConf.Le_RealFullChainPath,
			"/var/etc/acme-client/",
			vfsDB.List["acme-client"]+"/",
		)),
		vfsDB.MustReadFile(strings.ReplaceAll(
			interimIniConf.Le_RealKeyPath,
			"/var/etc/acme-client/",
			vfsDB.List["acme-client"]+"/",
		)),
	); {
	case err != nil:
		return nil, err
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

		//
		// production:
		//
		// leRealCertPath:      interimIniConf.Le_RealCertPath,
		// leRealCACertPath:    interimIniConf.Le_RealCACertPath,
		// leRealKeyPath:       interimIniConf.Le_RealKeyPath,
		// leRealFullChainPath: interimIniConf.Le_RealFullChainPath,
		cert: cert,

		//
		// MX
		//
		// mxList: io_net.LookupMX(
		// 	func() (outbound []string) {
		// 		switch {
		// 		case leconf[interimIniConf.Le_Domain].leAlt != nil:
		// 			outbound = leconf[interimIniConf.Le_Domain].leAlt
		// 		}
		// 		outbound = append(outbound, interimIniConf.Le_Domain)
		// 		return
		// 	}()),

		//
		// development:
		//
		leRealCertPath:      strings.ReplaceAll(interimIniConf.Le_RealCertPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/"),
		leRealCACertPath:    strings.ReplaceAll(interimIniConf.Le_RealCACertPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/"),
		leRealKeyPath:       strings.ReplaceAll(interimIniConf.Le_RealKeyPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/"),
		leRealFullChainPath: strings.ReplaceAll(interimIniConf.Le_RealFullChainPath, "/var/etc/acme-client/", vfsDB.List["acme-client"]+"/"),
	}

	return
}
