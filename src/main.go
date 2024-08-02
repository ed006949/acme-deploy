package main

import (
	"crypto/rsa"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
	"acme-deploy/src/l"
)

func main() {
	var (
		err       error
		xmlConfig = new(xmlConf)
	)

	switch err = xmlConfig.load(); {
	case errors.Is(err, l.ENOCONF):
		flag.PrintDefaults()
		l.Critical.E(err, nil)
	case err != nil:
		flag.PrintDefaults()
		l.Critical.E(err, nil)
	}

	// load data
	var (
		leConfigMap = func() (outbound map[string]*leConf) {
			outbound = make(map[string]*leConf)
			for _, b := range xmlConfig.ACMEClients {
				for _, d := range b.LEConf {
					for _, f := range append(d.LEAlt, d.LEDomain) {
						switch value, ok := outbound[f]; {
						case ok:
							l.Warning.E(l.EDUPDATA, l.F{"LE certificate": value.LEDomain})
							continue
						}
						outbound[f] = d
					}
				}
			}
			return
		}()
	)
	switch {
	case len(leConfigMap) == 0:
		l.Critical.E(l.ENOCONF, l.F{"message": "no ACME client config"})
	}

	for _, b := range xmlConfig.CGPs {
		b.Domains = make(map[string][]string)

		switch {
		case b.Token.URL == nil:
			b.Token.URL = &url.URL{
				Scheme: b.Token.Scheme,
				User:   url.UserPassword(b.Token.Username, b.Token.Password),
				Path:   b.Token.Path,
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
		b.Token.URL.Path = filepath.Join(b.Token.URL.Path) + "/" // CGP needs path separator at the end of the path

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
			l.Error.E(err, l.F{"message": "LISTDOMAINS", "CGP domain": b.Token.Name, "result": listDomains})
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
				l.Error.E(err, l.F{"message": "GETDOMAINALIASES", "CGP domain": b.Token.Name, "result": getDomainAliases})
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
							"message":        "UPDATEDOMAINSETTINGS",
							"LE certificate": value.LEDomain,
							"CGP domain":     c,
						})

						switch privateKey := value.Certificate.PrivateKey.(type) {
						case *rsa.PrivateKey:
							switch privateKey.Size() {
							case 512, 1024, 2048, 4096:
							default:
								l.Warning.E(io_crypto.EPrivKeySize, l.F{
									"message":        "CGP supports RSA only 512, 1024, 2048 or 4096 bits",
									"LE certificate": value.LEDomain,
									"CGP domain":     c,
									"type":           fmt.Sprintf("%T", value.Certificate.PrivateKey),
									"size":           privateKey.Size,
								})
								return
							}
						default:
							l.Warning.E(io_crypto.EPrivKeyType, l.F{
								"message":        "CGP supports only RSA and only 512, 1024, 2048 or 4096 bits",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"type":           fmt.Sprintf("%T", value.Certificate.PrivateKey),
							})
							return
						}

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
								"message":        "UPDATEDOMAINSETTINGS",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
							return
						case updateDomainSettings != nil:
							l.Warning.E(l.EUEDATA, l.F{
								"message":        "UPDATEDOMAINSETTINGS OK",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
						default:
							l.Informational.L(l.F{
								"message":        "UPDATEDOMAINSETTINGS OK",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
							})
						}

						return
					}
				}
			}()
		}
	}
}
