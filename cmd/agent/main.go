package main

import (
	"fmt"
	cl "github.com/c2pc/go-musthave-metrics/internal/client"
	"github.com/c2pc/go-musthave-metrics/internal/metric"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	waitTime       = 30 * time.Second
)

var (
	counterMetric metric.Metric[int64]
	gaugeMetric   metric.Metric[float64]
	client        cl.IClient
)

func main() {
	fmt.Println("Start metrics reporter")
	counterMetric = metric.NewCounterMetric()
	gaugeMetric = metric.NewGaugeMetric()

	client = cl.NewClient("http://localhost:8080")

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	waitTicker := time.NewTicker(waitTime)
	defer waitTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			pollMetrics()
		case <-reportTicker.C:
			reportMetrics()
		case <-waitTicker.C:
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
				return
			}
		}()
	}

	waitGroup.Wait()

	fmt.Println("Finish reporting metrics...")
}
