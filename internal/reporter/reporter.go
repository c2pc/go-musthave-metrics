package reporter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Updater interface {
	UpdateMetric(ctx context.Context, tp string, name string, value interface{}) error
}

type MetricReader[T float64 | int64] interface {
	GetName() string
	PollStats()
	GetStats() map[string]T
}

type Timer struct {
	PollInterval   int
	ReportInterval int
}

type Reporter struct {
	counterMetric MetricReader[int64]
	gaugeMetric   MetricReader[float64]
	client        Updater
	timer         Timer
}

func New(client Updater, timer Timer, counterMetric MetricReader[int64], gaugeMetric MetricReader[float64]) *Reporter {
	return &Reporter{
		counterMetric: counterMetric,
		gaugeMetric:   gaugeMetric,
		client:        client,
		timer:         timer,
	}
}

func (r *Reporter) Run(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(r.timer.PollInterval) * time.Second)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(time.Duration(r.timer.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			r.pollMetrics()
		case <-reportTicker.C:
			r.reportMetrics(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) pollMetrics() {
	fmt.Println("Starting polling metrics...")
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		r.counterMetric.PollStats()
	}()

	go func() {
		defer waitGroup.Done()
		r.gaugeMetric.PollStats()
	}()

	waitGroup.Wait()

	fmt.Println("Finish polling metrics...")
}

func (r *Reporter) reportMetrics(ctx context.Context) {
	fmt.Println("Starting reporting metrics...")
	waitGroup := sync.WaitGroup{}

	for key, value := range r.counterMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			fmt.Printf("Update Counter metric: %s = %v\n", key, value)
			err := r.client.UpdateMetric(ctx, r.counterMetric.GetName(), key, value)
			if err != nil {
				fmt.Printf("Error updating counter metric: %s = %v\n", key, err)
				return
			}
		}()
	}

	for key, value := range r.gaugeMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			fmt.Printf("Update Gauge metric: %s = %v\n", key, value)
			err := r.client.UpdateMetric(ctx, r.gaugeMetric.GetName(), key, value)
			if err != nil {
				fmt.Printf("Error updating gauge metric: %s = %v\n", key, err)
				return
			}
		}()
	}

	waitGroup.Wait()

	fmt.Println("Finish reporting metrics...")
}
