package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"mock-cmpp-stress-test/utils/log"
)

type TextMessages struct {
	Extend  string `toml:"extend"`
	Content string `toml:"content"`
	Phone   string `toml:"phone"`
}

type StressTestWorker struct {
	Name         string `toml:"name"`
	Concurrency  uint64 `toml:"concurrency"`
	DurationTime uint64 `toml:"duration_time"`
	TotalNum     uint64 `toml:"total_num"`
	Sleep        uint64 `toml:"sleep"`
}

type StressTestConfig struct {
	Enable   bool                `toml:"enable"`
	Workers  *[]StressTestWorker `toml:"workers"`
	Messages *[]TextMessages     `toml:"messages"`
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
