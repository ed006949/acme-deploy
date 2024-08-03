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

func (receiver *Token) command(payload string) (outbound []string, err error) {
	var (
		request  *http.Request
		response *http.Response
		interim  = *receiver.URL
		delim    = regexp.MustCompile(`[,\(\)]`)
		buffer   = new(bytes.Buffer)
	)

	interim.RawQuery = url.PathEscape(payload)

	switch request, err = http.NewRequest(http.MethodGet, interim.String(), nil); {
	case err != nil:
		return nil, err
	}

	// request.SetBasicAuth(receiver.Username, receiver.Password)

	switch response, err = http.DefaultClient.Do(request); {
	case err != nil:
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	switch {
	case response.StatusCode != 200:
		l.Error(l.Z{
			"":        l.EINVALRESPONSE,
			"message": response.Body,
		}) //
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
func (receiver *Token) Command(inbound *Command) (outbound []string, err error) {
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
					l.Debug(l.Z{"CGP server": receiver.Name, "payload len": len(payload)})
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

		l.Debug(l.Z{"CGP": receiver.Name, "payload": payload})
		switch outbound, err = receiver.command(payload); {
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

func (receiver *Command_Dictionary) compile() (outbound string) {
	outbound += "{"
	outbound += " "
	for a, b := range structs.Map(receiver) {
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

func (receiver *UPDATEDOMAINSETTINGS) compile() (outbound string) {
	outbound += "UPDATEDOMAINSETTINGS"
	outbound += " "
	outbound += receiver.DomainName
	outbound += " "
	outbound += receiver.NewSettings.compile()
	return
}

func (receiver *GETDOMAINALIASES) compile() (outbound string) {
	outbound += "GETDOMAINALIASES"
	outbound += " "
	outbound += receiver.DomainName
	return
}

func (receiver *MAINDOMAINNAME) compile() (outbound string) {
	outbound += "MAINDOMAINNAME"
	return
}
func (receiver *LISTDOMAINS) compile() (outbound string) {
	outbound += "LISTDOMAINS"
	return
}
