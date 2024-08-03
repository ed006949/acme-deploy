package io_xml

import (
	"encoding/xml"

	"acme-deploy/src/l"
)

func MustUnmarshal(data []byte, v interface{}) {
	switch err := xml.Unmarshal(data, &v); {
	case err != nil:
		l.Critical(l.Z{"": err})
	}
}
