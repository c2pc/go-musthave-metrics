package server

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddress = "localhost:8080"
	defaultStoreInterval = 300
	defaultRestore       = true
)

var (
	address         string
	storeInterval   int64
	fileStoragePath string
	restore         bool
	databaseDSN     string
)

type envConfig struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   string `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         string `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

type Config struct {
	Address         string
	StoreInterval   int64
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
}

func Parse() (*Config, error) {
	flag.StringVar(&address, "a", defaultServerAddress, "The Address of the server")
	flag.Int64Var(&storeInterval, "i", defaultStoreInterval, "The interval, in seconds, of the file store")
	flag.StringVar(&fileStoragePath, "f", "", "The path to the file storage")
	flag.BoolVar(&restore, "r", defaultRestore, "The restore flag")
	flag.StringVar(&databaseDSN, "d", "", "The database DSN")

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
		cfg.Address = address
	}

	//Parsing StoreInterval
	if envCfg.StoreInterval != "" {
		cfg.StoreInterval, err = strconv.ParseInt(envCfg.StoreInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse STORE_INTERVAL: %s", err)
		}
	} else if v := os.Getenv("STORE_INTERVAL"); v != "" {
		cfg.StoreInterval, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse STORE_INTERVAL: %s", err)
		}
	} else {
		cfg.StoreInterval = storeInterval
	}

	//Parsing FileStoragePath
	if envCfg.FileStoragePath != "" {
		cfg.FileStoragePath = envCfg.FileStoragePath
	} else if v := os.Getenv("ADDRESS"); v != "" {
		cfg.FileStoragePath = v
	} else {
		cfg.FileStoragePath = fileStoragePath
	}

	//Parsing Restore
	if envCfg.Restore != "" {
		cfg.Restore, err = strconv.ParseBool(envCfg.Restore)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RESTORE: %s", err)
		}
	} else if v := os.Getenv("RESTORE"); v != "" {
		cfg.Restore, err = strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RESTORE: %s", err)
		}
	} else {
		cfg.Restore = restore
	}

	//Parsing Database DSN
	if envCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = envCfg.DatabaseDSN
	} else if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DatabaseDSN = v
	} else {
		cfg.DatabaseDSN = databaseDSN
	}

	return cfg, nil
}
