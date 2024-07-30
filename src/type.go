package main

import (
	"encoding/xml"

	"acme-deploy/src/io_cgp"
	"acme-deploy/src/io_crypto"
)

type xmlConf struct {
	XMLName     xml.Name              `xml:"conf"`
	Daemon      *xmlConfDaemon        `xml:"daemon,omitempty"`
	ACMEClients []*xmlConfACMEClients `xml:"acme-clients>acme-client,omitempty"`
	CGPs        []*xmlConfCGPs        `xml:"cgps>cgp,omitempty"`
}

type xmlConfDaemon struct {
	Name      string `xml:"name,attr,omitempty"`
	Verbosity string `xml:"verbosity,attr,omitempty"`
	DryRun    bool   `xml:"dry-run,attr,omitempty"`
}

type xmlConfACMEClients struct {
	Name string `xml:"name,attr,omitempty"`
	Path string `xml:"path,attr,omitempty"`
}

type xmlConfCGPs struct {
	io_cgp.Token
	Domains map[string][]string `xml:"-"`
}

// type iniLEConf struct {
// 	leDomain            string   `ini:"Le_Domain"`
// 	leAlt               []string `ini:"Le_Alt" de//lim:","`
// 	leRealCertPath      string   `ini:"Le_RealCertPath"`
// 	leRealCACertPath    string   `ini:"Le_RealCACertPath"`
// 	leRealKeyPath       string   `ini:"Le_RealKeyPath"`
// 	leRealFullChainPath string   `ini:"Le_RealFullChainPath"`
// }

type iniLEConf struct {
	Le_Domain            string
	Le_Alt               []string `delim:","`
	Le_RealCertPath      string
	Le_RealCACertPath    string
	Le_RealKeyPath       string
	Le_RealFullChainPath string
}

type leConf struct {
	leDomain            string
	leAlt               []string
	leRealCertPath      string
	leRealCACertPath    string
	leRealKeyPath       string
	leRealFullChainPath string
	cert                *io_crypto.Certificate
	mxList              []string
}
