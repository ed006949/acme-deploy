package main

import (
	"errors"
	"flag"
	"io/fs"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/avfs/avfs"
	"github.com/avfs/avfs/idm/dummyidm"
	"github.com/avfs/avfs/vfs/memfs"
	"github.com/rs/zerolog"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_fs"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/io_xml"
	"acme-deploy/src/l"
)

func main() {
	_ = l.SetPackageVerbosity(zerolog.LevelInfoValue)

	var (
		err          error
		configFile   = flag.String("config", "", "xml config file")
		cliVerbosity = flag.String("verbosity", "", "verbosity level, overrides config")
		cliDryRun    = flag.Bool("dry-run", false, "dry-run, overrides config")
		xmlConfig    xmlConf
	)

	// parse CLI
	flag.Parse()
	switch {
	case !l.IsFlagExist("config"):
		flag.Usage()
		l.Critical.E(errors.New("config file is mandatory"), nil)
		fallthrough
	case l.IsFlagExist("verbosity"):
		_ = l.SetPackageVerbosity(*cliVerbosity)
		fallthrough
	case l.IsFlagExist("dry-run"):
		_ = l.SetPackageDryRun(*cliDryRun)
		fallthrough
	default:
	}

	// parse Config
	io_xml.MustUnmarshal(io_fs.MustReadFile(*configFile), &xmlConfig)
	_ = l.SetPackageVerbosity(xmlConfig.Daemon.Verbosity)
	_ = l.SetPackageDryRun(xmlConfig.Daemon.DryRun)
	_ = l.SetPackageName(xmlConfig.Daemon.Name)

	// re-parse CLI after Config, so CLI can override Config
	// _ = l.SetPackageDryRun(*cliDryRun && l.IsFlagExist("dry-run"))
	switch {
	case l.IsFlagExist("verbosity"):
		_ = l.SetPackageVerbosity(*cliVerbosity)
		fallthrough
	case l.IsFlagExist("dry-run"):
		_ = l.SetPackageDryRun(*cliDryRun)
		fallthrough
	default:
	}

	l.Notice.L(l.F{
		"config":    *configFile,
		"verbosity": l.PackageVerbosity.String(),
		"dry-run":   l.PackageDryRun,
	})

	// load data
	var (
		vfsDB = &io_vfs.VFSDB{
			List: map[string]string{},
			VFS: memfs.NewWithOptions(&memfs.Options{
				Idm:        dummyidm.NotImplementedIdm,
				User:       nil,
				Name:       "",
				OSType:     avfs.CurrentOSType(),
				SystemDirs: false,
			}),
		}

		leConfig    = make(map[string]*leConf)
		leConfigMap = make(map[string]*leConf)
		// MXList = make(map[string]struct{})
	)

	for _, b := range xmlConfig.ACMEClients {
		vfsDB.List[b.Name] = b.Path
	}

	vfsDB.MustReadVFS()
	// defer vfsDB.MustWriteVFS() // we don't change anything locally, yet

	// find acme.sh LE config files
	for _, b := range vfsDB.List {
		var (
			syncFn = func(name string, dirEntry fs.DirEntry, err error) error {
				switch {
				case err != nil:
					l.Critical.E(err, l.F{"name": name})
				}

				switch dirEntry.Type() {
				case fs.ModeDir:
				case fs.ModeSymlink:
				case 0:
					switch {
					case strings.HasSuffix(name, ".conf"):
						var (
							interimLEConf *leConf
						)

						switch interimLEConf, err = load(vfsDB, name); {
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

		vfsDB.MustWalkDir(b, syncFn)
	}

	for _, b := range leConfig {
		for _, d := range append(b.LEAlt, b.LEDomain) {
			switch value, ok := leConfigMap[d]; {
			case ok:
				l.Warning.E(errors.New("duplicate data"), l.F{"LE certificate": value.LEDomain})
				continue
			}
			leConfigMap[d] = b
		}
	}

	for _, b := range xmlConfig.CGPs {
		b.Domains = make(map[string][]string)

		switch {
		case b.Token.URL == nil:
			b.Token.URL = &url.URL{
				Scheme: b.Token.Scheme,
				User:   url.UserPassword(b.Token.Username, b.Token.Password),
				Path:   "/" + filepath.Join(b.Token.Path) + "/", // CGP needs path separator at the end of the path
				Host: func() string {
					switch b.Token.Port {
					case 0:
						return b.Token.Host
					default:
						return net.JoinHostPort(
							b.Token.Host,
							strconv.FormatUint(uint64(b.Token.Port), 10),
						)
					}
				}(),
			}
		}
		// }

		//

		// for _, b := range xmlConfig.CGPs {
		var (
			listDomains          []string
			getDomainAliases     []string
			updateDomainSettings []string
		)

		switch listDomains, err = b.Token.Command(
			&io_cgp.Command{
				Domain_Set_Administration: &io_cgp.Domain_Set_Administration{
					LISTDOMAINS: &io_cgp.LISTDOMAINS{},
				},
			},
		); {
		case err != nil:
			l.Error.E(err, l.F{"CGP domain": b.Token.Name, "result": listDomains})
			continue
		}

		for _, d := range listDomains {
			switch getDomainAliases, err = b.Token.Command(
				&io_cgp.Command{
					Domain_Administration: &io_cgp.Domain_Administration{
						GETDOMAINALIASES: &io_cgp.GETDOMAINALIASES{
							DomainName: d,
						},
					},
				},
			); {
			case err != nil:
				l.Error.E(err, l.F{"CGP domain": b.Token.Name, "result": getDomainAliases})
				continue
			}

			b.Domains[d] = getDomainAliases
		}
		// }

		//

		// for _, b := range xmlConfig.CGPs {
		for c, d := range b.Domains {
			func() {
				for _, h := range append(d, c) {
					switch value, ok := leConfigMap[h]; {
					case ok:
						l.Informational.L(l.F{"LE certificate": value.LEDomain, "CGP domain": c})
						switch updateDomainSettings, err = b.Command(
							&io_cgp.Command{
								Domain_Administration: &io_cgp.Domain_Administration{
									UPDATEDOMAINSETTINGS: &io_cgp.UPDATEDOMAINSETTINGS{
										DomainName: c,
										NewSettings: io_cgp.Command_Dictionary{
											CertificateType:   "YES",
											PrivateSecureKey:  string(value.Certificate.PrivateKeyRawPEM),
											SecureCertificate: string(value.Certificate.CertificatesRawPEM[0]),
											CAChain:           string(value.Certificate.CertificateCAChainRawPEM),
										},
									},
								},
							},
						); {
						case err != nil:
							l.Error.E(err, l.F{
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
							continue
						case updateDomainSettings != nil:
							l.Warning.E(errors.New("unexpected data"), l.F{
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
						}

						return
					}
				}
			}()
		}
	}

	//

}
