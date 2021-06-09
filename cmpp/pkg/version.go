package pkg

import cmpp "github.com/bigwhite/gocmpp"

const (
	V30            cmpp.Type = 0x30
	V21            cmpp.Type = 0x21
	V20            cmpp.Type = 0x20
	InvalidVersion cmpp.Type = 0x00
)

func String(v cmpp.Type) string {
	switch v {
	case V30:
		return "cmpp3.0"
	case V21:
		return "cmpp2.1"
	case V20:
		return "cmpp2.0"
	default:
	}
	return "unknown"
}

func GetVersion(version string) cmpp.Type {
	switch version {
	case "V30":
		return V30
	case "V21":
		return V21
	case "V20":
		return V20
	default:
		return InvalidVersion
	}
}
