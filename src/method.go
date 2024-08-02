package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/idm/dummyidm"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/go-ini/ini"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_fs"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/l"
)

func (receiver *leConf) loadEntry(vfsDB *io_vfs.VFSDB, name string) (err error) {
	var (
		interimLEConf leConf
		interimCert   *io_crypto.Certificate
		data          []byte
	)

	switch data, err = vfsDB.VFS.ReadFile(name); {
	case err != nil:
		return
	}

	data = bytes.ReplaceAll(data, []byte("/var/etc/acme-client/"), []byte(vfsDB.List["acme-client"]+"/"))

	switch err = ini.MapTo(&interimLEConf, data); {
	case err != nil:
		return
	case len(interimLEConf.LEDomain) == 0 ||
		len(interimLEConf.LERealCertPath) == 0 ||
		len(interimLEConf.LERealCACertPath) == 0 ||
		len(interimLEConf.LERealKeyPath) == 0 ||
		len(interimLEConf.LERealFullChainPath) == 0:
		return l.ENEDATA
	}

	switch interimCert, err = vfsDB.LoadX509KeyPair(
		interimLEConf.LERealFullChainPath,
		interimLEConf.LERealKeyPath,
	); {
	case err != nil:
		return
	}

	*receiver = leConf{
		LEDomain:            interimLEConf.LEDomain,
		LEAlt:               l.FilterSlice(interimLEConf.LEAlt, "no"), // OPNSense and acme.sh, alt domain name = "no" ????,
		LERealKeyPath:       interimLEConf.LERealKeyPath,
		LERealCertPath:      interimLEConf.LERealCertPath,
		LERealCACertPath:    interimLEConf.LERealCACertPath,
		LERealFullChainPath: interimLEConf.LERealFullChainPath,
		Certificate:         interimCert,
		// MXList:              io_net.LookupMX(append(receiver.LEAlt, interimLEConf.LEDomain)),
	}

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

func (receiver *xmlConf) getConfig() (err error) {
	var (
		vfsDB = &io_vfs.VFSDB{
			List: make(map[string]string),
			VFS: memfs.NewWithOptions(&memfs.Options{
				Idm:        dummyidm.NotImplementedIdm,
				User:       nil,
				Name:       "",
				OSType:     avfs.CurrentOSType(),
				SystemDirs: false,
			}),
		}

		cliConfig    = flag.String(l.PackageFlagConfig, os.Getenv("ACME_DEPLOY_CONFIG"), "xml config file")
		cliVerbosity = flag.String(l.PackageFlagVerbosity, os.Getenv("ACME_DEPLOY_VERBOSITY"), "verbosity level, overrides config")
		cliDryRun    = flag.Bool(l.PackageFlagDryRun, l.StripErr1(l.ParseBool(os.Getenv("ACME_DEPLOY_DRYRUN"))), "dry run, overrides config")

		// cliLEHome          = flag.String("home", os.Getenv("ACME_HOME_DIR"), "ACME_HOME_DIR")
		// cliLECertHome      = flag.String("cert-home", os.Getenv("ACME_CERT_HOME_DIR"), "ACME_CERT_HOME_DIR")
		// cliLECertPath      = flag.String("certpath", os.Getenv("ACME_CERT_FILE"), "ACME_CERT_FILE")
		// cliLEKeyPath       = flag.String("keypath", os.Getenv("ACME_KEY_FILE"), "ACME_KEY_FILE")
		// cliLECAPath        = flag.String("capath", os.Getenv("ACME_CHAIN_FILE"), "ACME_CHAIN_FILE")
		// cliLEFullChainPath = flag.String("fullchainpath", os.Getenv("ACME_FULLCHAIN_FILE"), "ACME_FULLCHAIN_FILE")

		cliCGP = flag.String("cgp", os.Getenv("ACME_DEPLOY_CGP"), "CGP HTTPU URL")
	)

	flag.Parse()

	switch {
	case len(flag.Args()) == 5: // acme.sh deploy
		var (
			args = flag.Args()[:1]
		)
		for _, name := range flag.Args()[1:] {
			switch name, err = filepath.Abs(name); {
			case err != nil:
				return
			}
			args = append(args, name)
		}

		var (
			xmlConfig = &xmlConf{
				Daemon: &xmlConfDaemon{
					Name:      _c_deploy,
					Verbosity: l.PackageVerbosity.String(),
					DryRun:    l.PackageDryRun,
				},
				ACMEClients: []*xmlConfACMEClients{
					{
						Name: _c_deploy,
						Path: "",
						LEConf: map[string]*leConf{
							flag.Arg(0): {
								LEDomain:            args[0],
								LEAlt:               nil,
								LERealKeyPath:       args[1],
								LERealCertPath:      args[2],
								LERealCACertPath:    args[3],
								LERealFullChainPath: args[4],
								Certificate:         nil,
							},
						},
					}},
				CGPs: []*xmlConfCGPs{{Token: &io_cgp.Token{Name: _c_deploy}}},
			}
		)
		for _, name := range args[1:] {
			switch err = vfsDB.CopyFromFS2VFS(name); {
			case err != nil:
				return
			}
		}

		switch xmlConfig.CGPs[0].URL, err = l.UrlParse(*cliCGP); {
		case err != nil:
			return
		}

		switch xmlConfig.ACMEClients[0].LEConf[args[0]].Certificate, err = vfsDB.LoadX509KeyPair(args[4], args[1]); {
		case err != nil:
			return
		}

		xmlConfig.ACMEClients[0].LEConf[args[0]].LEAlt = l.FilterSlice(
			xmlConfig.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].DNSNames,
			xmlConfig.ACMEClients[0].LEConf[args[0]].LEDomain,
		)

		l.Informational.L(l.F{"message": _c_deploy, "CGP": xmlConfig.CGPs[0].URL.Redacted()})
		l.Informational.L(l.F{"message": _c_deploy, "LEDomain": xmlConfig.ACMEClients[0].LEConf[args[0]].LEDomain})
		l.Informational.L(l.F{"message": _c_deploy, "LEAlt": xmlConfig.ACMEClients[0].LEConf[args[0]].LEAlt})
		l.Informational.L(l.F{"message": _c_deploy, "CN": xmlConfig.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].Subject.CommonName})

		*receiver = *xmlConfig
		return nil

	case len(*cliConfig) != 0: // CLI
		switch {
		case l.FlagIsFlagExist(l.PackageFlagVerbosity):
			_ = l.SetPackageVerbosity(*cliVerbosity)
			fallthrough
		case l.FlagIsFlagExist(l.PackageFlagDryRun):
			_ = l.SetPackageDryRun(*cliDryRun)
			fallthrough
		default:
		}

		var (
			cliConfigFile string
			data          []byte
			xmlConfig     = new(xmlConf)
		)

		switch cliConfigFile, err = filepath.Abs(*cliConfig); {
		case err != nil:
			return
		}
		switch err = vfsDB.CopyFromFS2VFS(cliConfigFile); {
		case err != nil:
			return
		}
		switch data, err = vfsDB.VFS.ReadFile(cliConfigFile); {
		case err != nil:
			return
		}
		switch err = xml.Unmarshal(data, xmlConfig); {
		case err != nil:
			return
		case &xmlConfig == nil:
			return l.ENOCONF
		}

		_ = l.SetPackageVerbosity(xmlConfig.Daemon.Verbosity)
		_ = l.SetPackageDryRun(xmlConfig.Daemon.DryRun)
		_ = l.SetPackageName(xmlConfig.Daemon.Name)
		switch {
		case l.FlagIsFlagExist(l.PackageFlagVerbosity):
			_ = l.SetPackageVerbosity(*cliVerbosity)
			fallthrough
		case l.FlagIsFlagExist(l.PackageFlagDryRun):
			_ = l.SetPackageDryRun(*cliDryRun)
			fallthrough
		default:
		}

		for _, b := range xmlConfig.ACMEClients {
			vfsDB.List[b.Name] = b.Path
			b.LEConf = make(map[string]*leConf)
		}
		switch err = vfsDB.CopyFS2VFS(); {
		case err != nil:
			return
		}

		// find acme.sh LE config files
		for _, b := range xmlConfig.ACMEClients {
			for _, d := range vfsDB.List {
				var (
					findLEConf = func(name string, dirEntry fs.DirEntry, err error) (fnErr error) {
						switch {
						case err != nil:
							return err
						}

						switch dirEntry.Type() {
						case fs.ModeDir:
						case fs.ModeSymlink:
						case 0:
							switch {
							case strings.HasSuffix(name, ".csr.conf"): // skip CSR config files
							case strings.HasSuffix(name, ".conf"):
								var (
									interimLEConf = new(leConf)
								)

								switch err = interimLEConf.loadEntry(vfsDB, name); {
								case errors.Is(err, l.ENODATA):
									l.Debug.E(err, l.F{"file": name})
								case errors.Is(err, l.ENEDATA):
									l.Debug.E(err, l.F{"file": name})
								case err != nil:
									l.Warning.E(err, l.F{"file": name})
								default:
									b.LEConf[interimLEConf.LEDomain] = interimLEConf
								}
							}
						default:
						}

						return nil
					}
				)

				switch err = vfsDB.VFS.WalkDir(d, findLEConf); {
				case err != nil:
					return
				}
			}
		}

		*receiver = *xmlConfig
		return nil

	default:
		return l.ENOCONF
	}
}
