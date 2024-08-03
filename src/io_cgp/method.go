package io_cgp

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/fatih/structs"

	"acme-deploy/src/l"
)

func (r *Token) command(payload string) (outbound []string, err error) {
	var (
		request  *http.Request
		response *http.Response
		interim  = *r.URL
		delim    = regexp.MustCompile(`[,\(\)]`)
		buffer   = new(bytes.Buffer)
	)

	interim.RawQuery = url.PathEscape(payload)

	switch request, err = http.NewRequest(http.MethodGet, interim.String(), nil); {
	case err != nil:
		return nil, err
	}

	// request.SetBasicAuth(r.Username, r.Password)

	switch response, err = http.DefaultClient.Do(request); {
	case err != nil:
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	switch {
	case response.StatusCode != 200:
		l.Z{l.E: l.EINVALRESPONSE, l.M: response.Body}.Error()
		return nil, l.EINVALRESPONSE
	}

	switch _, err = buffer.ReadFrom(response.Body); {
	case err != nil:
		return nil, err
	}

	for _, b := range strings.Fields(buffer.String()) {
		for _, d := range delim.Split(b, -1) {
			switch {
			case len(d) == 0:
				continue
			}
			outbound = append(outbound, d)
		}
	}

	return
}

// Command will execute only first command found
func (r *Token) Command(inbound *Command) (outbound []string, err error) {
	var (
		payload       string
		emptyResponse bool // check if response must be empty
	)

	switch {
	case inbound != nil:
		payload += "command"
		payload += "="

		switch {
		case inbound.Domain_Administration != nil:

			switch {
			case inbound.Domain_Administration.GETDOMAINALIASES != nil:
				payload += inbound.Domain_Administration.GETDOMAINALIASES.compile()

			case inbound.Domain_Administration.UPDATEDOMAINSETTINGS != nil:
				emptyResponse = true
				payload += inbound.Domain_Administration.UPDATEDOMAINSETTINGS.compile()

				switch {
				case l.DryRun.Value():
					l.Z{"CGP server": r.Name, "payload len": len(payload)}.Debug()
					payload = ""
				}

			default:
				return nil, EComSetDomAdm
			}

		case inbound.Domain_Set_Administration != nil:

			switch {
			case inbound.Domain_Set_Administration.MAINDOMAINNAME != nil:
				payload += inbound.Domain_Set_Administration.MAINDOMAINNAME.compile()

			case inbound.Domain_Set_Administration.LISTDOMAINS != nil:
				payload += inbound.Domain_Set_Administration.LISTDOMAINS.compile()

			default:
				return nil, EComSetDomSetAdm
			}

		default:
			return nil, EComSet
		}

		l.Z{"CGP": r.Name, "payload": payload}.Debug()
		switch outbound, err = r.command(payload); {
		case err != nil:
			return
		case emptyResponse && outbound != nil:
			return outbound, l.EINVALRESPONSE
		default:
			return
		}

	default:
		return nil, ECom
	}

}

func (r *Command_Dictionary) compile() (outbound string) {
	outbound += "{"
	outbound += " "
	for a, b := range structs.Map(r) {
		outbound += a
		switch {
		case len(b.(string)) > 0:
			outbound += "="
			switch a {
			case "CAChain", "PrivateSecureKey", "SecureCertificate":
				outbound += "["
				outbound += b.(string)
				outbound += "]"
			default:
				outbound += b.(string)
			}
		}
		outbound += ";"
		outbound += " "
	}
	outbound += " "
	outbound += "}"
	return
}

func (r *UPDATEDOMAINSETTINGS) compile() (outbound string) {
	outbound += "UPDATEDOMAINSETTINGS"
	outbound += " "
	outbound += r.DomainName
	outbound += " "
	outbound += r.NewSettings.compile()
	return
}

func (r *GETDOMAINALIASES) compile() (outbound string) {
	outbound += "GETDOMAINALIASES"
	outbound += " "
	outbound += r.DomainName
	return
}

func (r *MAINDOMAINNAME) compile() (outbound string) {
	outbound += "MAINDOMAINNAME"
	return
}
func (r *LISTDOMAINS) compile() (outbound string) {
	outbound += "LISTDOMAINS"
	return
}
