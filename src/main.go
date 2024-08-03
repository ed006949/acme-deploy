package main

import (
	"crypto/rsa"
	"errors"
	"flag"

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
		l.Z{l.E: err}.Critical()
	case err != nil:
		flag.PrintDefaults()
		l.Z{l.E: err}.Critical()
	}

	for _, b := range xmlConfig.CGPs {
		var (
			listDomains []string
		)
		switch listDomains, err = b.Token.Command(
			&io_cgp.Command{
				Domain_Set_Administration: &io_cgp.Domain_Set_Administration{
					LISTDOMAINS: &io_cgp.LISTDOMAINS{},
				},
			},
		); {
		case err != nil:
			continue
		}

		for _, d := range listDomains {
			var (
				getDomainAliases []string
			)
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
				continue
			}
			b.Domains[d] = getDomainAliases
		}

		for c, d := range b.Domains {
			func() {
				for _, h := range append(d, c) {
					switch value, ok := xmlConfig.LEConfMap[h]; {
					case ok:
						l.Z{"LE certificate": value.LEDomain}.Debug()

						switch privateKey := value.Certificate.PrivateKey.(type) {
						case *rsa.PrivateKey:
							switch privateKey.Size() {
							case 512, 1024, 2048, 4096:
							default:
								l.Z{l.E: io_crypto.EPrivKeySize, l.M: "CGP supports RSA only 512, 1024, 2048 or 4096 bits", "LE certificate": value.LEDomain, l.T: privateKey, "size": privateKey.Size}.Warning()
								return
							}
						default:
							l.Z{l.E: io_crypto.EPrivKeyType, l.M: "CGP supports only RSA and only 512, 1024, 2048 or 4096 bits", "LE certificate": value.LEDomain, l.T: privateKey}.Warning()
							return
						}

						var (
							updateDomainSettings []string
						)
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
							// return
						case updateDomainSettings != nil:
						default:
						}
					}
				}
			}()
		}
	}
}
