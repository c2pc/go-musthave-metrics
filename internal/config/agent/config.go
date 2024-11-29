package agent

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

var (
	serverAddress  string
	pollInterval   int
	reportInterval int
	hashKey        string
)

type envConfig struct {
	ServerAddress  string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	HashKey        string `env:"KEY"`
}

type Config struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
	HashKey        string
}

func Parse() (*Config, error) {
	flag.StringVar(&serverAddress, "a", defaultServerAddress, "The Address of the server")
	flag.IntVar(&pollInterval, "p", defaultPollInterval, "The interval between polls in seconds")
	flag.IntVar(&reportInterval, "r", defaultReportInterval, "The interval between reports in seconds")
	flag.StringVar(&hashKey, "k", "", "The hash key")

	cfg := Config{}

	flag.Parse()
	if len(flag.Args()) > 0 {
		return nil, fmt.Errorf("unknown argument: %s", flag.Args()[0])
	}

	envCfg := envConfig{}
	err := env.Parse(&envCfg)
	if err != nil {
		return nil, err
	}

	if envCfg.ServerAddress != "" {
		cfg.ServerAddress = envCfg.ServerAddress
	} else if address := os.Getenv("ADDRESS"); address != "" {
		cfg.ServerAddress = address
	} else {
		cfg.ServerAddress = serverAddress
	}

	if envCfg.PollInterval != 0 {
		cfg.PollInterval = envCfg.PollInterval
	} else {
		cfg.PollInterval = pollInterval
	}

	if envCfg.ReportInterval != 0 {
		cfg.ReportInterval = envCfg.ReportInterval
	} else {
		cfg.ReportInterval = reportInterval
	}

	//Parsing Hash Key
	if envCfg.HashKey != "" {
		cfg.HashKey = envCfg.HashKey
	} else if v := os.Getenv("KEY"); v != "" {
		cfg.HashKey = v
	} else {
		cfg.HashKey = hashKey
	}

	return &cfg, nil
}
