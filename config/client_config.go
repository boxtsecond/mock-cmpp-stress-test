package config

type CmppAccount struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Ip       string `toml:"ip"`
	Port     uint16 `toml:"port"`
	SpID     string `toml:"sp_id"`
	SpCode   string `toml:"sp_code"`
}

type CmppClientConfig struct {
	Version            string         `toml:"version"`
	TimeOut            uint           `toml:"read_timeout"`
	Retries            uint           `toml:"retries"`
	ActiveTestInterval uint           `toml:"active_test_interval"`
	MaxNoRespPkgNum    uint           `toml:"max_no_resp_pkg_num"`
	Enable             bool           `toml:"enable"`
	Accounts           *[]CmppAccount `toml:"accounts"`
}
