package reporter

import (
	"context"
	"sync"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/model"
)

type Updater interface {
	UpdateMetric(ctx context.Context, tp string, name string, value interface{}) (*model.Metrics, error)
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

	for key, value := range r.counterMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			metric, err := r.client.UpdateMetric(ctx, r.counterMetric.GetName(), key, value)
			if err != nil {
				logger.Log.Info("Error updating counter metric", logger.Any("key", key), logger.Error(err))
				return
			}
			logger.Log.Info("Update Counter metric", logger.Any("key", key), logger.Any("value", value), logger.Any("response", metric))
		}()
	}

	for key, value := range r.gaugeMetric.GetStats() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			metric, err := r.client.UpdateMetric(ctx, r.gaugeMetric.GetName(), key, value)
			if err != nil {
				logger.Log.Info("Error updating gauge metric", logger.Any("key", key), logger.Error(err))
				return
			}
			logger.Log.Info("Update Gauge metric", logger.Any("key", key), logger.Any("value", value), logger.Any("response", metric))
		}()
	}

	waitGroup.Wait()

	logger.Log.Info("Finish reporting metrics...")
}
