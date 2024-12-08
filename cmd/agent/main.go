package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	cl "github.com/c2pc/go-musthave-metrics/internal/client"
	config "github.com/c2pc/go-musthave-metrics/internal/config/agent"
	"github.com/c2pc/go-musthave-metrics/internal/hash"
	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/metric"
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
	"github.com/c2pc/go-musthave-metrics/internal/worker_pool"
)

type Reporter interface {
	Run(context.Context)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := logger.Initialize("info")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v\n", err)
	}
	defer logger.Log.Sync()

	logger.Log.Info("Starting metrics reporter")
	defer logger.Log.Info("Stopping metrics reporter")

	cfg, err := config.Parse()
	if err != nil {
		logger.Log.Fatal("failed to parse config", logger.Error(err))
	}

	counterMetric := metric.NewCounterMetric()
	gaugeMetric := metric.NewGaugeMetric()

	var client reporter.Updater
	if cfg.HashKey != "" {
		hasher, err := hash.New(cfg.HashKey)
		if err != nil {
			logger.Log.Error("failed to initialize hasher", logger.Error(err))
			return
		}
		client = cl.NewClient(cfg.ServerAddress, hasher)
	} else {
		client = cl.NewClient(cfg.ServerAddress, nil)
	}

	workerPool := worker_pool.New(ctx, cfg.RateLimit)
	defer workerPool.ShutDown()

	var report Reporter = reporter.New(client, workerPool, reporter.Timer{
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
	}, counterMetric, gaugeMetric)

	go report.Run(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-ctx.Done():
			return
		case <-quit:
			return
		}
	}
}
