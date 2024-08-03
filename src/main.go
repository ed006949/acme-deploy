package main

import (
	"crypto/rsa"
	"errors"
	"flag"
	"fmt"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
	"acme-deploy/src/l"
)

func main() {
	var (
		err       error
		xmlConfig = new(xmlConf)
	)

	l.Informational(l.Z{
		"version": "1.0",
		"title":   "acme-deploy",
		"license": "Apache-2.0",
		"message": "mEssAgE",
		"":        l.EINVAL,
	})

	switch err = xmlConfig.load(); {
	case errors.Is(err, l.ENOCONF):
		flag.PrintDefaults()
		l.Critical(l.Z{"": err})
	case err != nil:
		flag.PrintDefaults()
		l.Critical(l.Z{"": err})
	}

	for _, b := range xmlConfig.CGPs {
		var (
			listDomains          []string
			getDomainAliases     []string
			updateDomainSettings []string
		)

		l.Debug(l.Z{
			"message":    "LISTDOMAINS",
			"CGP server": b.Token.Name,
		})
		switch listDomains, err = b.Token.Command(
			&io_cgp.Command{
				Domain_Set_Administration: &io_cgp.Domain_Set_Administration{
					LISTDOMAINS: &io_cgp.LISTDOMAINS{},
				},
			},
		); {
		case err != nil:
			l.Error(l.Z{"": err,
				"message":    "LISTDOMAINS",
				"CGP server": b.Token.Name,
				"result":     listDomains,
			})
			continue
		}
		l.Informational(l.Z{
			"message":    "LISTDOMAINS OK",
			"CGP server": b.Token.Name,
			"result":     len(listDomains),
		})

		for _, d := range listDomains {
			l.Debug(l.Z{
				"message":    "GETDOMAINALIASES",
				"CGP domain": b.Token.Name,
			})
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
				l.Error(l.Z{"": err,
					"message":    "GETDOMAINALIASES",
					"CGP domain": b.Token.Name,
					"result":     getDomainAliases,
				})
				continue
			}

			b.Domains[d] = getDomainAliases
			l.Informational(l.Z{
				"message":    "GETDOMAINALIASES OK",
				"CGP domain": b.Token.Name,
				"result":     len(getDomainAliases),
			})
		}

		for c, d := range b.Domains {
			func() {
				for _, h := range append(d, c) {
					switch value, ok := xmlConfig.LEConfMap[h]; {
					case ok:
						l.Debug(l.Z{
							"message":        "UPDATEDOMAINSETTINGS",
							"LE certificate": value.LEDomain,
							"CGP domain":     c,
						})

						switch privateKey := value.Certificate.PrivateKey.(type) {
						case *rsa.PrivateKey:
							switch privateKey.Size() {
							case 512, 1024, 2048, 4096:
							default:
								l.Warning(l.Z{"": io_crypto.EPrivKeySize,
									"message":        "CGP supports RSA only 512, 1024, 2048 or 4096 bits",
									"LE certificate": value.LEDomain,
									"CGP domain":     c,
									"type":           fmt.Sprintf("%T", value.Certificate.PrivateKey),
									"size":           privateKey.Size,
								})
								return
							}
						default:
							l.Warning(l.Z{"": io_crypto.EPrivKeyType,
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
							l.Error(l.Z{"": err,
								"message":        "UPDATEDOMAINSETTINGS",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
							return
						case updateDomainSettings != nil:
							l.Warning(l.Z{"": l.EUEDATA,
								"message":        "UPDATEDOMAINSETTINGS OK",
								"LE certificate": value.LEDomain,
								"CGP domain":     c,
								"result":         updateDomainSettings,
							})
						default:
							l.Informational(l.Z{
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
