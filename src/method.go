package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
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
	"github.com/go-ini/ini"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_vfs"
	"acme-deploy/src/l"
)

func (r *leConf) load(vfsDB *io_vfs.VFSDB, name string) (err error) {
	var (
		data []byte
	)

	switch data, err = vfsDB.VFS.ReadFile(name); {
	case err != nil:
		return
	}

	data = bytes.ReplaceAll(data, []byte("/var/etc/acme-client/"), []byte(vfsDB.List["acme-client"]+"/"))

	switch err = ini.MapTo(r, data); {
	case err != nil:
		return
	case len(r.LEDomain) == 0 ||
		len(r.LERealCertPath) == 0 ||
		len(r.LERealCACertPath) == 0 ||
		len(r.LERealKeyPath) == 0 ||
		len(r.LERealFullChainPath) == 0:
		return l.ENEDATA
	}

	switch r.Certificate, err = vfsDB.LoadX509KeyPair(
		r.LERealFullChainPath,
		r.LERealKeyPath,
	); {
	case err != nil:
		return
	}

	r.LEAlt = l.FilterSlice(r.LEAlt, "no") // OPNSense and acme.sh, alt domain name = "no" ????

	return
}

func (r *xmlConf) load() (err error) {
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

		// cliLEHome          = flag.String("home", os.Getenv("ACME_HOME_DIR"), "ACME_HOME_DIR")
		// cliLECertHome      = flag.String("cert-home", os.Getenv("ACME_CERT_HOME_DIR"), "ACME_CERT_HOME_DIR")
		// cliLECertPath      = flag.String("certpath", os.Getenv("ACME_CERT_FILE"), "ACME_CERT_FILE")
		// cliLEKeyPath       = flag.String("keypath", os.Getenv("ACME_KEY_FILE"), "ACME_KEY_FILE")
		// cliLECAPath        = flag.String("capath", os.Getenv("ACME_CHAIN_FILE"), "ACME_CHAIN_FILE")
		// cliLEFullChainPath = flag.String("fullchainpath", os.Getenv("ACME_FULLCHAIN_FILE"), "ACME_FULLCHAIN_FILE")

		cliCGP = flag.String("cgp", os.Getenv(l.EnvName("CGP")), "CGP HTTPU URL ("+l.EnvName("CGP")+")")
	)
	l.InitCLI()

	r.LEConfMap = make(map[string]*leConf)

	switch {
	case len(flag.Args()) == 5:
		l.Deploy.Set()

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

		r.Daemon = &l.ControlType{
			Name:      l.Name.Value(),
			Verbosity: l.Verbosity.Value(),
			DryRun:    l.DryRun.Value(),
		}
		r.ACMEClients = []*xmlConfACMEClients{
			{
				Name: l.Mode.String(),
				Path: "",
				LEConf: map[string]*leConf{
					args[0]: {
						LEDomain:            args[0],
						LEAlt:               nil,
						LERealKeyPath:       args[1],
						LERealCertPath:      args[2],
						LERealCACertPath:    args[3],
						LERealFullChainPath: args[4],
						Certificate:         nil,
					},
				},
			}}
		r.CGPs = []*xmlConfCGPs{{
			Token: &io_cgp.Token{
				Name: l.Mode.String(),
			},
		}}

		switch r.CGPs[0].URL, err = l.UrlParse(*cliCGP); {
		case err != nil:
			return
		}

		for _, name := range args[1:] {
			switch err = vfsDB.CopyFromFS(name); {
			case err != nil:
				return
			}
		}

		switch r.ACMEClients[0].LEConf[args[0]].Certificate, err = vfsDB.LoadX509KeyPair(args[4], args[1]); {
		case err != nil:
			return
		}

		r.ACMEClients[0].LEConf[args[0]].LEAlt = l.FilterSlice(
			r.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].DNSNames,
			r.ACMEClients[0].LEConf[args[0]].LEDomain,
		)

		l.Z{l.M: l.Mode.String(), "CGP": r.CGPs[0].URL.Redacted()}.Informational()
		l.Z{l.M: l.Mode.String(), "LEDomain": r.ACMEClients[0].LEConf[args[0]].LEDomain}.Informational()
		l.Z{l.M: l.Mode.String(), "LEAlt": r.ACMEClients[0].LEConf[args[0]].LEAlt}.Informational()
		l.Z{l.M: l.Mode.String(), "CN": r.ACMEClients[0].LEConf[args[0]].Certificate.Certificates[0].Subject.CommonName}.Informational()

	case len(l.Config.String()) != 0: //
		l.CLI.Set()

		var (
			cliConfigFile string
			data          []byte
		)

		switch cliConfigFile, err = filepath.Abs(l.Config.String()); {
		case err != nil:
			return
		}
		switch err = vfsDB.CopyFromFS(cliConfigFile); {
		case err != nil:
			return
		}
		switch data, err = vfsDB.VFS.ReadFile(cliConfigFile); {
		case err != nil:
			return
		}
		switch err = xml.Unmarshal(data, r); {
		case err != nil:
			return
		}

		for _, b := range r.ACMEClients {
			vfsDB.List[b.Name] = b.Path
			b.LEConf = make(map[string]*leConf)
		}
		switch err = vfsDB.LoadFromFS(); {
		case err != nil:
			return
		}

		// find acme.sh LE config files
		for _, b := range r.ACMEClients {
			for _, d := range vfsDB.List {
				var (
					findLEConf = func(name string, dirEntry fs.DirEntry, fnErr error) (err error) {
						switch {
						case fnErr != nil:
							return fnErr
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
								case errors.Is(err, l.ENODATA):
									return nil
								case errors.Is(err, l.ENEDATA):
									return nil
								case err != nil:
									return
								}

								b.LEConf[interimLEConf.LEDomain] = interimLEConf

							}
						default:
						}

						return
					}
				)

				switch err = vfsDB.VFS.WalkDir(d, findLEConf); {
				case err != nil:
					return
				}
			}
		}

	default:
		return l.ENOCONF
	}

	for _, b := range r.CGPs {
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
	}

	for _, b := range r.ACMEClients {
		for _, d := range b.LEConf {
			for _, f := range append(d.LEAlt, d.LEDomain) {
				switch value, ok := r.LEConfMap[f]; {
				case ok:
					l.Z{l.E: l.EDUPDATA, "LE certificate": value.LEDomain}.Warning()
					continue
				}
				r.LEConfMap[f] = d
			}
		}
	}

	switch {
	case len(r.LEConfMap) == 0:
		return l.ENEDATA
	}

	return
}
