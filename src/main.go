package main

import (
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"strconv"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/l"
)

func main() {
	var (
		xmlConfig *xmlConf
	)

	switch xmlConfig, err = getConfig(); {
	case errors.Is(err, l.ENOCONF):
		l.Critical.E(err, nil)
	case err != nil:
		l.Critical.E(err, nil)
	}

	// load data
	var (
		leConfig    = make(map[string]*leConf)
		leConfigMap = make(map[string]*leConf)
		// MXList = make(map[string]struct{})
	)

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
}
