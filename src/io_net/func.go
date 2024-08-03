package io_net

import (
	"errors"
	"net"

	"acme-deploy/src/l"
)

func LookupMX(names []string) (outbound []string) {
	for _, name := range names {
		var (
			interim, err = net.LookupMX(name)
			errDetail    *net.DNSError
			_            = errors.As(err, &errDetail)
		)

		switch {
		case errDetail != nil && errDetail.IsNotFound:
			continue
		case err != nil:
			l.Z{l.E: err, l.M: "MX", "name": name}.Warning()
			continue
		}
		for _, b := range interim {
			outbound = append(outbound, b.Host)
		}
	}
	return
}
