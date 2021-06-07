package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"mock-cmpp-stress-test/utils/log"
)

type StressTestConfig struct {
	Concurrency  uint `toml:"concurrency"`
	DurationTime uint `toml:"duration_time"`
	TotalNum     uint `toml:"total_num"`
}

type RedisConfig struct {
	IP       string `toml:"ip"`
	Port     uint16 `toml:"port"`
	Password string `toml:"password"`
	TimeOut  uint   `toml:"timeout"`
	Enable   bool   `toml:"enable"`
	Wait     bool   `toml:"wait"`
}

type Config struct {
	ClientConfig *CmppClientConfig `toml:"cmpp_client"`
	ServerConfig *CmppServerConfig `toml:"cmpp_server"`
	StressTest   *StressTestConfig `toml:"stress_test"`
	Log          *log.Config       `toml:"log"`
	Redis        *RedisConfig      `toml:"redis"`
}

var ConfigObj Config

func Init() error {
	defaultCfgFile := "./config/config.toml"

	cfgFile := flag.String("c", defaultCfgFile, "config file")
	flag.Parse()

	if _, err := toml.DecodeFile(*cfgFile, &ConfigObj); err != nil {
		return err
	}

	return nil
}
