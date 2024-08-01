package main

import (
	"bytes"
	"encoding/xml"

	"github.com/go-ini/ini"

	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_fs"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/l"
)

func (receiver *leConf) loadEntry(vfsDB *io_vfs.VFSDB, name string) (err error) {
	// switch err = ini.MapTo(&interimIniConf, vfsDB.MustReadFile(name)); {
	switch err = ini.MapTo(receiver, bytes.ReplaceAll(
		vfsDB.MustReadFile(name),
		[]byte("/var/etc/acme-client/"),
		[]byte(vfsDB.List["acme-client"]+"/"),
	)); {
	case err != nil:
		return
	case len(receiver.LEDomain) == 0 || len(receiver.LERealCertPath) == 0 || len(receiver.LERealCACertPath) == 0 || len(receiver.LERealKeyPath) == 0 || len(receiver.LERealFullChainPath) == 0:
		return l.ENEDATA
	}

	receiver.LEAlt = l.FilterSlice(receiver.LEAlt, "no") // OPNSense and acme.sh, alt domain name = "no" ????

	switch receiver.Certificate, err = io_crypto.X509KeyPair(
		vfsDB.MustReadFile(receiver.LERealFullChainPath),
		vfsDB.MustReadFile(receiver.LERealKeyPath),
	); {
	case err != nil:
		return
	}

	// receiver.MXList = io_net.LookupMX(append(receiver.LEAlt, interimIniConf.LEDomain))

	return
}

func (receiver *xmlConf) load(vfsDB *io_vfs.VFSDB, name string) (err error) {
	switch err = xml.Unmarshal(vfsDB.MustReadFile(io_fs.MustAbs(name)), receiver); {
	case err != nil:
		return
	}

	_ = l.SetPackageVerbosity(receiver.Daemon.Verbosity)
	_ = l.SetPackageDryRun(receiver.Daemon.DryRun)
	_ = l.SetPackageName(receiver.Daemon.Name)

	return
}
func (receiver *xmlConf) mustLoad(vfsDB *io_vfs.VFSDB, name string) {
	switch err := receiver.load(vfsDB, name); {
	case err != nil:
		l.Critical.E(err, nil)
	}
}
