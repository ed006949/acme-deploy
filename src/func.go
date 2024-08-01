package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/idm/dummyidm"
	"github.com/avfs/avfs/vfs/memfs"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/l"
)

func getConfig() (outbound *xmlConf, err error) {
	var (
		xmlConfig *xmlConf
		vfsDB     = &io_vfs.VFSDB{
			List: make(map[string]string),
			VFS: memfs.NewWithOptions(&memfs.Options{
				Idm:        dummyidm.NotImplementedIdm,
				User:       nil,
				Name:       "",
				OSType:     avfs.CurrentOSType(),
				SystemDirs: false,
			}),
		}

		cliConfig    = flag.String("config", os.Getenv("ACME_DEPLOY_CONFIG"), "xml config file")
		cliVerbosity = flag.String("verbosity", os.Getenv("ACME_DEPLOY_VERBOSITY"), "verbosity level, overrides config")
		cliDryRun    = flag.Bool("dryrun", l.StripErr1(l.ParseBool(os.Getenv("ACME_DEPLOY_DRYRUN"))), "dryrun, overrides config")

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
			interimXMLConfig = &xmlConf{
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
			outboundChain []byte
			outboundKey   []byte
		)

		for _, name := range args[1:] {
			switch err = vfsDB.CopyFromFS2VFS(name); {
			case err != nil:
				return
			}
		}

		switch interimXMLConfig.CGPs[0].URL, err = url.Parse(*cliCGP); {
		case err != nil:
			return
		case len(interimXMLConfig.CGPs[0].URL.String()) == 0:
			return nil, l.ENEDATA
		}

		switch outboundChain, err = vfsDB.VFS.ReadFile(args[4]); {
		case err != nil:
			return
		}
		switch outboundKey, err = vfsDB.VFS.ReadFile(args[1]); {
		case err != nil:
			return
		}
		switch interimXMLConfig.ACMEClients[0].LEConf[args[0]].Certificate, err = io_crypto.X509KeyPair(outboundChain, outboundKey); {
		case err != nil:
			return
		}

		interimXMLConfig.ACMEClients[0].LEConf[args[0]].LEAlt = l.FilterSlice(
			interimXMLConfig.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].DNSNames,
			interimXMLConfig.ACMEClients[0].LEConf[args[0]].LEDomain,
		)

		l.Informational.L(l.F{"message": _c_deploy, "CGP": interimXMLConfig.CGPs[0].URL.Redacted()})
		l.Informational.L(l.F{"message": _c_deploy, "LEDomain": interimXMLConfig.ACMEClients[0].LEConf[args[0]].LEDomain})
		l.Informational.L(l.F{"message": _c_deploy, "LEAlt": interimXMLConfig.ACMEClients[0].LEConf[args[0]].LEAlt})
		l.Informational.L(l.F{"message": _c_deploy, "CN": interimXMLConfig.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].Subject.CommonName})

		return interimXMLConfig, nil

	case l.IsFlagExist("config"): // CLI _c_cli
		_ = l.SetPackageVerbosity(*cliVerbosity)
		_ = l.SetPackageDryRun(*cliDryRun)

		var (
			cliConfigFile = *cliConfig
			data          []byte
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
		}

		_ = l.SetPackageVerbosity(xmlConfig.Daemon.Verbosity)
		_ = l.SetPackageDryRun(xmlConfig.Daemon.DryRun)
		_ = l.SetPackageName(xmlConfig.Daemon.Name)
		_ = l.SetPackageVerbosity(*cliVerbosity)
		_ = l.SetPackageDryRun(*cliDryRun)

		for _, b := range xmlConfig.ACMEClients {
			vfsDB.List[b.Name] = b.Path
		}

		// find acme.sh LE config files
		for _, b := range vfsDB.List {
			var (
				findLEConf = func(name string, dirEntry fs.DirEntry, err error) error {
					switch {
					case err != nil:
						l.Critical.E(err, l.F{"name": name})
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
							case errors.Is(err, l.ENEDATA):
								l.Debug.E(err, l.F{"file": name})
							case err != nil:
								l.Warning.E(err, l.F{"file": name})
							default:
								leConfig[interimLEConf.LEDomain] = interimLEConf
							}
						}
					default:
					}

					return nil
				}
			)

			vfsDB.MustWalkDir(b, findLEConf)
		}

	default:
		return nil, l.ENOCONF
	}
}
