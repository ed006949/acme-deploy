package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
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
		err       error
		xmlConfig xmlConf

		cliConfig    = flag.String("config", os.Getenv("ACME_DEPLOY_CONFIG"), "xml config file")
		cliVerbosity = flag.String("verbosity", os.Getenv("ACME_DEPLOY_VERBOSITY"), "verbosity level, overrides config")
		cliDryRun    = flag.Bool("dryrun", l.StripErr1(l.ParseBool(os.Getenv("ACME_DEPLOY_DRYRUN"))), "dryrun, overrides config")

		// cliLEHome          = flag.String("home", os.Getenv("ACME_HOME_DIR"), "ACME_HOME_DIR")
		// cliLECertHome      = flag.String("cert-home", os.Getenv("ACME_CERT_HOME_DIR"), "ACME_CERT_HOME_DIR")
		// cliLECertPath      = flag.String("certpath", os.Getenv("ACME_CERT_FILE"), "ACME_CERT_FILE")
		// cliLEKeyPath       = flag.String("keypath", os.Getenv("ACME_KEY_FILE"), "ACME_KEY_FILE")
		// cliLECAPath        = flag.String("capath", os.Getenv("ACME_CHAIN_FILE"), "ACME_CHAIN_FILE")
		// cliLEFullChainPath = flag.String("fullchainpath", os.Getenv("ACME_FULLCHAIN_FILE"), "ACME_FULLCHAIN_FILE")
	)
	//
	//        $this->acme_env['DEPLOY_PROXMOXVE_USER'] = (string)$this->config->acme_proxmoxve_user;
	//        $this->acme_env['DEPLOY_PROXMOXVE_SERVER'] = (string)$this->config->acme_proxmoxve_server;
	//        $this->acme_env['DEPLOY_PROXMOXVE_SERVER_PORT'] = (string)$this->config->acme_proxmoxve_port;
	//        $this->acme_env['DEPLOY_PROXMOXVE_NODE_NAME'] = (string)$this->config->acme_proxmoxve_nodename;
	//        $this->acme_env['DEPLOY_PROXMOXVE_USER_REALM'] = (string)$this->config->acme_proxmoxve_realm;
	//        $this->acme_env['DEPLOY_PROXMOXVE_API_TOKEN_NAME'] = (string)$this->config->acme_proxmoxve_tokenid;
	//        $this->acme_env['DEPLOY_PROXMOXVE_API_TOKEN_KEY'] = (string)$this->config->acme_proxmoxve_tokenkey;
	//
	//        $this->acme_args[] = LeUtils::execSafe('--home %s', self::ACME_HOME_DIR);
	//        $this->acme_args[] = LeUtils::execSafe('--cert-home %s', sprintf(self::ACME_CERT_HOME_DIR, $this->cert_id));
	//        $this->acme_args[] = LeUtils::execSafe('--certpath %s', sprintf(self::ACME_CERT_FILE, $this->cert_id));
	//        $this->acme_args[] = LeUtils::execSafe('--keypath %s', sprintf(self::ACME_KEY_FILE, $this->cert_id));
	//        $this->acme_args[] = LeUtils::execSafe('--capath %s', sprintf(self::ACME_CHAIN_FILE, $this->cert_id));
	//        $this->acme_args[] = LeUtils::execSafe('--fullchainpath %s', sprintf(self::ACME_FULLCHAIN_FILE, $this->cert_id));
	//
	//        _cdomain="$1"
	//        _ckey="$2"
	//        _ccert="$3"
	//        _cca="$4"
	//        _cfullchain="$5"
	//

	// parse CLI
	flag.Parse()

	switch {
	case len(flag.Args()) == 5: // acme.sh deploy
		l.Notice.L(l.F{
			"args": fmt.Sprintf("%s", flag.Args()),
		})
		// xmlConfig = xmlConf{
		// 	Daemon: &xmlConfDaemon{
		// 		Name:      "acme-deploy-CLI",
		// 		Verbosity: "info",
		// 		DryRun:    true,
		// 	},
		// 	ACMEClients: []*xmlConfACMEClients{
		// 		{
		// 			Name: "CLI",
		// 			Path: "",
		// 		},
		// 	},
		// 	CGPs: nil,
		// }
	}

	switch {
	case !l.IsFlagExist("config"):
		flag.Usage()
		l.Critical.E(l.ENOCONF, nil)
		fallthrough
	case l.IsFlagExist("verbosity"):
		_ = l.SetPackageVerbosity(*cliVerbosity)
		fallthrough
	case l.IsFlagExist("dryrun"):
		_ = l.SetPackageDryRun(*cliDryRun)
		fallthrough
	default:
	}

	// parse Config
	io_xml.MustUnmarshal(io_fs.MustReadFile(*cliConfig), &xmlConfig)
	_ = l.SetPackageVerbosity(xmlConfig.Daemon.Verbosity)
	_ = l.SetPackageDryRun(xmlConfig.Daemon.DryRun)
	_ = l.SetPackageName(xmlConfig.Daemon.Name)

	// re-parse CLI after Config, so CLI can override Config
	// _ = l.SetPackageDryRun(*cliDryRun && l.IsFlagExist("dryrun"))
	switch {
	case l.IsFlagExist("verbosity"):
		_ = l.SetPackageVerbosity(*cliVerbosity)
		fallthrough
	case l.IsFlagExist("dryrun"):
		_ = l.SetPackageDryRun(*cliDryRun)
		fallthrough
	default:
	}

	l.Notice.L(l.F{
		"config":    *cliConfig,
		"verbosity": l.PackageVerbosity.String(),
		"dryrun":    l.PackageDryRun,
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

						switch err = interimLEConf.load(vfsDB, name); {
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

	for _, b := range leConfig {
		for _, d := range append(b.LEAlt, b.LEDomain) {
			switch value, ok := leConfigMap[d]; {
			case ok:
				l.Warning.E(l.EDUPDATA, l.F{"LE certificate": value.LEDomain})
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
						l.Informational.L(l.F{
							"LE certificate": value.LEDomain,
							"CGP domain":     c,
							"message":        "update",
						})
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
							return
						// case updateDomainSettings != nil:
						// 	l.Warning.E(l.EUEDATA, l.F{
						// 		"LE certificate": value.LEDomain,
						// 		"CGP domain":     c,
						// 		"result":         updateDomainSettings,
						// 	})
						default:
							l.Informational.L(l.F{
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"message":        "update OK",
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
