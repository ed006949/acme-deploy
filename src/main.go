package main

import (
	"crypto/rsa"
	"errors"
	"flag"
	"fmt"

	"github.com/rs/zerolog"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
	"acme-deploy/src/l"
)

func main() {
	var (
		err       error
		xmlConfig = new(xmlConf)
	)

	l.Z{
		"version":                "1.0",
		"title":                  "acme-deploy",
		"license":                "Apache-2.0",
		zerolog.MessageFieldName: "mEssAgE",
		zerolog.ErrorFieldName:   l.EINVAL,
	}.Informational()

	switch err = xmlConfig.load(); {
	case errors.Is(err, l.ENOCONF):
		flag.PrintDefaults()
		l.Z{l.E: err}.Critical()
	case err != nil:
		flag.PrintDefaults()
		l.Z{l.E: err}.Critical()
	}

	for _, b := range xmlConfig.CGPs {
		var (
			listDomains          []string
			getDomainAliases     []string
			updateDomainSettings []string
		)

		l.Z{l.M: "LISTDOMAINS", "CGP server": b.Token.Name}.Debug()
		switch listDomains, err = b.Token.Command(
			&io_cgp.Command{
				Domain_Set_Administration: &io_cgp.Domain_Set_Administration{
					LISTDOMAINS: &io_cgp.LISTDOMAINS{},
				},
			},
		); {
		case err != nil:
			l.Z{l.E: err, l.M: "LISTDOMAINS", "CGP server": b.Token.Name, "result": listDomains}.Error()
			continue
		}
		l.Z{l.M: "LISTDOMAINS OK", "CGP server": b.Token.Name, "result": len(listDomains)}.Informational()

		for _, d := range listDomains {
			l.Z{l.M: "GETDOMAINALIASES", "CGP domain": b.Token.Name}.Debug()
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
				l.Z{l.E: err, l.M: "GETDOMAINALIASES", "CGP domain": b.Token.Name, "result": getDomainAliases}.Error()
				continue
			}

			b.Domains[d] = getDomainAliases
			l.Z{l.M: "GETDOMAINALIASES OK", "CGP domain": b.Token.Name, "result": len(getDomainAliases)}.Informational()
		}

		for c, d := range b.Domains {
			func() {
				for _, h := range append(d, c) {
					switch value, ok := xmlConfig.LEConfMap[h]; {
					case ok:
						l.Z{l.M: "UPDATEDOMAINSETTINGS", "LE certificate": value.LEDomain, "CGP domain": c}.Debug()

						switch privateKey := value.Certificate.PrivateKey.(type) {
						case *rsa.PrivateKey:
							switch privateKey.Size() {
							case 512, 1024, 2048, 4096:
							default:
								l.Z{l.E: io_crypto.EPrivKeySize, l.M: "CGP supports RSA only 512, 1024, 2048 or 4096 bits", "LE certificate": value.LEDomain, "CGP domain": c, "type": fmt.Sprintf("%T", value.Certificate.PrivateKey), "size": privateKey.Size}.Warning()
								return
							}
						default:
							l.Z{l.E: io_crypto.EPrivKeyType, l.M: "CGP supports only RSA and only 512, 1024, 2048 or 4096 bits", "LE certificate": value.LEDomain, "CGP domain": c, "type": fmt.Sprintf("%T", value.Certificate.PrivateKey)}.Warning()
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
							l.Z{l.E: err, l.M: "UPDATEDOMAINSETTINGS", "LE certificate": value.LEDomain, "CGP domain": c, "result": updateDomainSettings}.Error()
							return
						case updateDomainSettings != nil:
							l.Z{l.E: l.EUEDATA, l.M: "UPDATEDOMAINSETTINGS OK", "LE certificate": value.LEDomain, "CGP domain": c, "result": updateDomainSettings}.Warning()
						default:
							l.Z{l.M: "UPDATEDOMAINSETTINGS OK", "LE certificate": value.LEDomain, "CGP domain": c}.Informational()
						}

						return
					}
				}
			}()
		}
	}
}
