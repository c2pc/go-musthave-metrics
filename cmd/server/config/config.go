package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress   = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "tmp/metric.tmp"
	defaultRestore         = true
)

var (
	address         = flag.String("a", defaultServerAddress, "The Address of the server")
	storeInterval   = flag.Int64("i", defaultStoreInterval, "The interval, in seconds, of the file store")
	fileStoragePath = flag.String("f", defaultFileStoragePath, "The path to the file storage")
	restore         = flag.Bool("r", defaultRestore, "The restore flag")
)

type envConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   string `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         string `env:"RESTORE"`
}

type Config struct {
	Address         string
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
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

	//Parsing Address
	if envCfg.Address != "" {
		cfg.Address = envCfg.Address
	} else if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Address = v
	} else {
		cfg.Address = *address
	}

	//Parsing StoreInterval
	if envCfg.StoreInterval != "" {
		cfg.StoreInterval, err = strconv.ParseInt(envCfg.StoreInterval, 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse STORE_INTERVAL: %s", err))
		}
	} else if v := os.Getenv("STORE_INTERVAL"); v != "" {
		cfg.StoreInterval, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse STORE_INTERVAL: %s", err))
		}
	} else {
		cfg.StoreInterval = *storeInterval
	}

	//Parsing FileStoragePath
	if envCfg.FileStoragePath != "" {
		cfg.FileStoragePath = envCfg.FileStoragePath
	} else if v := os.Getenv("ADDRESS"); v != "" {
		cfg.FileStoragePath = v
	} else {
		cfg.FileStoragePath = *fileStoragePath
	}

	//Parsing Restore
	if envCfg.Restore != "" {
		cfg.Restore, err = strconv.ParseBool(envCfg.Restore)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse RESTORE: %s", err))
		}
	} else if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Restore, err = strconv.ParseBool(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse RESTORE: %s", err))
		}
	} else {
		cfg.Restore = *restore
	}

	return cfg, nil
}
