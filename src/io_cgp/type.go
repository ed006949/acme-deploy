package io_cgp

import (
	"net/url"
)

type Token struct {
	Name     string   `xml:"name,attr,omitempty"`
	URL      *url.URL `xml:"url,attr,omitempty"`
	Scheme   string   `xml:"scheme,attr,omitempty"`
	Username string   `xml:"username,attr,omitempty"`
	Password string   `xml:"password,attr,omitempty"`
	Host     string   `xml:"host,attr,omitempty"`
	Port     uint16   `xml:"port,attr,omitempty"`
	Path     string   `xml:"path,attr,omitempty"`
}

type Command struct {
	*Domain_Set_Administration
	*Domain_Administration
}

type Domain_Set_Administration struct {
	*MAINDOMAINNAME
	*LISTDOMAINS
}

type Domain_Administration struct {
	*GETDOMAINALIASES
	*UPDATEDOMAINSETTINGS
}

type MAINDOMAINNAME struct{}
type LISTDOMAINS struct{}

type GETDOMAINALIASES struct {
	DomainName string
}

type Command_Dictionary struct {
	CAChain           string
	CertificateType   string
	PrivateSecureKey  string
	SecureCertificate string
}

type UPDATEDOMAINSETTINGS struct {
	DomainName  string
	NewSettings Command_Dictionary
}
