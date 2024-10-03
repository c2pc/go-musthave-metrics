package main

import (
	"context"
	"fmt"
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/c2pc/go-musthave-metrics/cmd/agent/config"
	cl "github.com/c2pc/go-musthave-metrics/internal/client"
	"github.com/c2pc/go-musthave-metrics/internal/metric"
)

type Reporter interface {
	Run(context.Context)
}

func main() {
	fmt.Println("Starting metrics reporter")
	defer fmt.Println("Stopping metrics reporter")

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
		return
	}

	counterMetric := metric.NewCounterMetric()
	gaugeMetric := metric.NewGaugeMetric()

	client := cl.NewClient(cfg.ServerAddress)

	var report Reporter = reporter.New(client, reporter.Timer{
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
	}, counterMetric, gaugeMetric)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.WaitTime)*time.Second)
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
