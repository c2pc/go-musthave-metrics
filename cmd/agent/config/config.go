package config

import (
	"flag"
	"fmt"
)

const (
	defaultServerAddress  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

var (
	serverAddress  = flag.String("a", defaultServerAddress, "The Address of the server")
	pollInterval   = flag.Int("p", defaultPollInterval, "The interval between reports in seconds")
	reportInterval = flag.Int("r", defaultReportInterval, "The interval between reports in seconds")
)

type Config struct {
	ServerAddress  string
	PollInterval   int
	ReportInterval int
	WaitTime       int
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.Parse()

	if len(flag.Args()) > 0 {
		return nil, fmt.Errorf("unknown argument: %s", flag.Args()[0])
	}

	cfg.ServerAddress = *serverAddress
	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.WaitTime = 30

	return cfg, nil
}
