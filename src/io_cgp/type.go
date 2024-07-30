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
	Domain_Set_Administration *Domain_Set_Administration `structs:",omitempty"`
	Domain_Administration     *Domain_Administration     `structs:",omitempty"`
}

type Domain_Set_Administration struct {
	MAINDOMAINNAME *MAINDOMAINNAME `structs:",omitempty"`
	LISTDOMAINS    *LISTDOMAINS    `structs:",omitempty"`
}

type Domain_Administration struct {
	GETDOMAINALIASES     *GETDOMAINALIASES     `structs:",omitempty"`
	UPDATEDOMAINSETTINGS *UPDATEDOMAINSETTINGS `structs:",omitempty"`
}

type MAINDOMAINNAME struct{}
type LISTDOMAINS struct{}

type GETDOMAINALIASES struct {
	DomainName string `structs:",omitempty"`
}

type Command_Dictionary struct {
	CAChain           string `structs:",omitempty"`
	CertificateType   string `structs:",omitempty"`
	PrivateSecureKey  string `structs:",omitempty"`
	SecureCertificate string `structs:",omitempty"`
}

type UPDATEDOMAINSETTINGS struct {
	DomainName  string             `structs:",omitempty"`
	NewSettings Command_Dictionary `structs:",omitempty"`
}
