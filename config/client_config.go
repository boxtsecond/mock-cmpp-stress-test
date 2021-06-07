package config

type CmppAccount struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Ip       string `toml:"ip"`
	Port     uint16 `toml:"port"`
}

type CmppMessages struct {
	SpID    string `toml:"sp_id"`
	SpCode  string `toml:"sp_code"`
	Extend  string `toml:"extend"`
	Content string `toml:"content"`
	Phone   string `toml:"phone"`
}

type CmppClientConfig struct {
	Key                string          `toml:"key"`
	Version            string          `toml:"version"`
	TimeOut            uint            `toml:"read_timeout"`
	ActiveTestInterval uint            `toml:"active_test_interval"`
	MaxNoRespPkgNum    uint            `toml:"max_no_resp_pkg_num"`
	Enable             bool            `toml:"enable"`
	Accounts           *[]CmppAccount  `toml:"accounts"`
	Messages           *[]CmppMessages `toml:"messages"`
}
