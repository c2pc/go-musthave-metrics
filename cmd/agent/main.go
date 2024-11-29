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
)

type Reporter interface {
	Run(context.Context)
}

func main() {
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
		hasher := hash.New(cfg.HashKey)
		client = cl.NewClient(cfg.ServerAddress, hasher)
	} else {
		client = cl.NewClient(cfg.ServerAddress, nil)
	}

	var report Reporter = reporter.New(client, reporter.Timer{
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
	}, counterMetric, gaugeMetric)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
