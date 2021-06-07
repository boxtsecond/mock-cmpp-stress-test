package pkg

type Version uint8

const (
	V30            Version = 0x30
	V21            Version = 0x21
	V20            Version = 0x20
	InvalidVersion Version = 0x00
)

func (v Version) String() string {
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

func GetVersion(version string) Version {
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
