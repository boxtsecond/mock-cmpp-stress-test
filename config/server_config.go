package config

type CmppServerAuth struct {
	UserName string `toml:"username"`
	Password string `toml:"password"`
	SpId     string `toml:"sp_id"`
	SpCode   string `toml:"sp_code"`
}

type CmppServerConfig struct {
	IP              string            `toml:"ip"`
	Port            uint16            `toml:"port"`
	Enable          bool              `toml:"enable"`
	Version         string            `toml:"version"`
	HeartBeat       uint8             `toml:"heartbeat"`
	MaxNoRspPkgs    uint              `toml:"max_no_resp_pkgs"`
	Auths           *[]CmppServerAuth `toml:"auths"`
	DeliverInterval uint8             `toml:"deliver_interval"` // 回执发送间隔时间
}
