package reporter

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/c2pc/go-musthave-metrics/internal/retry"
)

type Updater interface {
	UpdateMetric(ctx context.Context, metrics []model.Metrics) error
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
	logger.Log.Info("Starting polling metrics...")
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

	logger.Log.Info("Finish polling metrics...")
}

func (r *Reporter) reportMetrics(ctx context.Context) {
	logger.Log.Info("Starting reporting metrics...")
	waitGroup := sync.WaitGroup{}

	var counters []model.Metrics
	for key, value := range r.counterMetric.GetStats() {
		counters = append(counters, model.Metrics{
			ID:    key,
			Type:  r.counterMetric.GetName(),
			Delta: &value,
		})

	}

	var gauges []model.Metrics
	for key, value := range r.gaugeMetric.GetStats() {
		gauges = append(gauges, model.Metrics{
			ID:    key,
			Type:  r.gaugeMetric.GetName(),
			Value: &value,
		})
	}

	if len(counters) > 0 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := r.updateMetrics(ctx, counters)
			if err != nil {
				logger.Log.Info("Error updating counters metric", logger.Error(err))
				return
			}
		}()
	}

	if len(gauges) > 0 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := r.updateMetrics(ctx, gauges)
			if err != nil {
				logger.Log.Info("Error updating gauge metric", logger.Error(err))
				return
			}
		}()
	}

	waitGroup.Wait()

	logger.Log.Info("Finish reporting metrics...")
}

func (r *Reporter) updateMetrics(ctx context.Context, metrics []model.Metrics) error {
	return retry.Retry(
		func() error {
			return r.client.UpdateMetric(ctx, metrics)
		},
		func(err error) bool {
			var netErr net.Error
			if errors.As(err, &netErr) || errors.Is(err, context.DeadlineExceeded) {
				return true
			}
			return false
		},
		[]time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	)
}
