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
	LEConfMap   map[string]*leConf    `xml:"-"`
}

type xmlConfDaemon struct {
	Name      string `xml:"name,attr,omitempty"`
	Verbosity string `xml:"verbosity,attr,omitempty"`
	DryRun    bool   `xml:"dryrun,attr,omitempty"`
}

type xmlConfACMEClients struct {
	Name   string             `xml:"name,attr,omitempty"`
	Path   string             `xml:"path,attr,omitempty"`
	LEConf map[string]*leConf `xml:"-"`
}

type xmlConfCGPs struct {
	*io_cgp.Token
	Domains map[string][]string `xml:"-"`
}

type leConf struct {
	LEDomain            string                 `ini:"Le_Domain"`            // _cdomain="$1"
	LEAlt               []string               `ini:"Le_Alt" delim:","`     //
	LERealKeyPath       string                 `ini:"Le_RealKeyPath"`       // _ckey="$2"
	LERealCertPath      string                 `ini:"Le_RealCertPath"`      // _ccert="$3"
	LERealCACertPath    string                 `ini:"Le_RealCACertPath"`    // _cca="$4"
	LERealFullChainPath string                 `ini:"Le_RealFullChainPath"` // _cfullchain="$5"
	Certificate         *io_crypto.Certificate `ini:"-"`                    //
	MXList              []string               `ini:"-"`                    //
}
