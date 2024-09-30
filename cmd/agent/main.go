package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/c2pc/go-musthave-metrics/cmd/agent/config"
	cl "github.com/c2pc/go-musthave-metrics/internal/client"
	"github.com/c2pc/go-musthave-metrics/internal/metric"
)

const (
	waitTime = 30 * time.Second
)

var (
	counterMetric metric.Metric[int64]
	gaugeMetric   metric.Metric[float64]
	client        cl.IClient
)

func main() {
	fmt.Println("Start metrics reporter")

	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
		return
	}

	counterMetric = metric.NewCounterMetric()
	gaugeMetric = metric.NewGaugeMetric()

	client = cl.NewClient(cfg.ServerAddress)

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			pollMetrics()
		case <-reportTicker.C:
			reportMetrics()
		case <-time.After(waitTime):
			fmt.Println("Finish metrics reporter")
			return
		}
	}
}

func pollMetrics() {
	fmt.Println("Start polling metrics...")
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		counterMetric.PollStats()
	}()

	go func() {
		defer waitGroup.Done()
		gaugeMetric.PollStats()
	}()

	waitGroup.Wait()

	fmt.Println("Finish polling metrics...")
}

func reportMetrics() {
	fmt.Println("Start reporting metrics...")
	waitGroup := sync.WaitGroup{}

	for key, value := range counterMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			fmt.Printf("Update Counter metric: %s = %v\n", key, value)
			err := client.UpdateMetric(counterMetric.GetName(), string(key), value)
			if err != nil {
				fmt.Printf("Error updating counter metric: %s = %v\n", key, err)
				return
			}
		}()
	}

	for key, value := range gaugeMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			fmt.Printf("Update Gauge metric: %s = %v\n", key, value)
			err := client.UpdateMetric(gaugeMetric.GetName(), string(key), value)
			if err != nil {
				fmt.Printf("Error updating gauge metric: %s = %v\n", key, err)
				return
			}
		}()
	}

	waitGroup.Wait()

	fmt.Println("Finish reporting metrics...")
}
