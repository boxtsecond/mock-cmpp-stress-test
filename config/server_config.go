package config

type CmppServerConfig struct {
	IP      string `toml:"ip"`
	Port    uint16 `toml:"port"`
	Enable  bool   `toml:"enable"`
	Version string `toml:"version"`
}
