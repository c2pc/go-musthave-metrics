package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/c2pc/go-musthave-metrics/cmd/agent/config"
	cl "github.com/c2pc/go-musthave-metrics/internal/client"
	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/metric"
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
	"go.uber.org/zap"
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
		logger.Log.Fatal("failed to parse config", zap.Error(err))
		return
	}

	counterMetric := metric.NewCounterMetric()
	gaugeMetric := metric.NewGaugeMetric()

	client := cl.NewClient(cfg.ServerAddress)

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
