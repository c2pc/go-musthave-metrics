package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/c2pc/go-musthave-metrics/internal/config/server"
	"github.com/c2pc/go-musthave-metrics/internal/database"
	"github.com/c2pc/go-musthave-metrics/internal/database/migrate"
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/server"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/c2pc/go-musthave-metrics/internal/sync"
)

func main() {
	err := logger.Initialize("info")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v\n", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Starting Server APP")
	defer logger.Log.Info("Shutting Down Server APP")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Parse()
	if err != nil {
		logger.Log.Fatal("failed to parse config", logger.Error(err))
	}

	var db *database.DB
	var memoryType storage.Type
	if cfg.DatabaseDSN == "" {
		memoryType = storage.TypeMemory
	} else {
		memoryType = storage.TypeDB
		db, err = database.New(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("failed to connect to database", logger.Error(err))
		}
		defer db.Close()

		purl, err := url.Parse(cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("failed to parse database URL", logger.Error(err))
		}

		err = migrate.Migrate(db.DB, purl.Path)
		if err != nil {
			logger.Log.Fatal("failed to migrate database", logger.Error(err))
		}
	}

	gaugeStorage, err := storage.NewGaugeStorage(memoryType, db)
	if err != nil {
		logger.Log.Fatal("failed to initialize gaugeStorage", logger.Error(err))
	}
	counterStorage, err := storage.NewCounterStorage(memoryType, db)
	if err != nil {
		logger.Log.Fatal("failed to initialize counterStorage", logger.Error(err))
	}

	if cfg.FileStoragePath != "" && cfg.DatabaseDSN == "" {
		syncer, err := sync.Start(ctx, sync.Config{
			StoreInterval:   cfg.StoreInterval,
			FileStoragePath: cfg.FileStoragePath,
			Restore:         cfg.Restore,
		}, gaugeStorage, counterStorage)
		if err != nil {
			logger.Log.Fatal("failed to start syncer", logger.Error(err))
		}
		defer syncer.Close()
	}

	handlers := handler.NewHandler(gaugeStorage, counterStorage, db)

	httpServer := server.NewServer(handlers, cfg.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		logger.Log.Info("Starting Server", logger.Any("address", cfg.Address))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Info("Error to ListenAndServe", logger.Error(err))
			close(quit)
		}
	}()

	<-quit

	const timeout = 5 * time.Second
	ctx2, shutdown := context.WithTimeout(ctx, timeout)
	defer shutdown()

	if err := httpServer.Stop(ctx2); err != nil {
		logger.Log.Info("Failed to Stop Server", logger.Error(err))
	}
}
