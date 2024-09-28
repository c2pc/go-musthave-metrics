package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"os"
)

const (
	defaultServerAddress = "localhost:8080"
)

var (
	address = flag.String("a", defaultServerAddress, "The Address of the server")
)

type envConfig struct {
	Address string `env:"ADDRESS"`
}

type Config struct {
	Address string
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.Parse()

	if len(flag.Args()) > 0 {
		return nil, fmt.Errorf("unknown argument: %s", flag.Args()[0])
	}

	envCfg := envConfig{}
	err := env.Parse(&envCfg)
	if err != nil {
		return nil, err
	}

	if envCfg.Address != "" {
		cfg.Address = envCfg.Address
	} else if addr := os.Getenv("ADDRESS"); addr != "" {
		cfg.Address = addr
	} else {
		cfg.Address = *address
	}

	return cfg, nil
}
