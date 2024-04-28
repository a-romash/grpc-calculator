package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string           `yaml:"env"`
	CountCalcs int              `yaml:"count_calculators"`
	GRPCClient GRPCClientConfig `yaml:"grpc_client"`
	Durations  DurationsConfig  `yaml:"durations"`
}

type GRPCClientConfig struct {
	Addr         string `yaml:"orch_addr"`
	RetriesCount int    `yaml:"retries_count"`
}

type DurationsConfig struct {
	Plus  time.Duration `yaml:"plus"`
	Minus time.Duration `yaml:"minus"`
	Mult  time.Duration `yaml:"mult"`
	Del   time.Duration `yaml:"del"`
	Pow   time.Duration `yaml:"pow"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
