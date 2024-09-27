package config

import (
	"flag"
	"fmt"
)

const (
	defaultServerAddress = "localhost:8080"
)

var (
	address = flag.String("a", defaultServerAddress, "The Address of the server")
)

type Config struct {
	Address string
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.Parse()

	if len(flag.Args()) > 0 {
		return nil, fmt.Errorf("unknown argument: %s", flag.Args()[0])
	}

	cfg.Address = *address

	return cfg, nil
}
