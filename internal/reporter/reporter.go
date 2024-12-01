package reporter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/model"
	"github.com/c2pc/go-musthave-metrics/internal/retry"
)

type Updater interface {
	UpdateMetric(ctx context.Context, metrics ...model.Metric) error
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
	rateLimit     int
	jobs          chan []model.Metric
	results       chan error
}

func New(client Updater, timer Timer, counterMetric MetricReader[int64], gaugeMetric MetricReader[float64], rateLimit int) *Reporter {
	if rateLimit <= 0 {
		rateLimit = 1
	}

	return &Reporter{
		counterMetric: counterMetric,
		gaugeMetric:   gaugeMetric,
		client:        client,
		timer:         timer,
		rateLimit:     rateLimit,
		jobs:          make(chan []model.Metric),
		results:       make(chan error),
	}
}

func (r *Reporter) Run(ctx context.Context) {
	for i := 1; i <= r.rateLimit; i++ {
		go r.worker(ctx, i, r.jobs, r.results)
	}

	go r.poll(ctx)
	go r.report(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) poll(ctx context.Context) {
	pollTicker := time.NewTicker(time.Duration(r.timer.PollInterval) * time.Second)
	defer pollTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			r.pollMetrics()
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) report(ctx context.Context) {
	reportTicker := time.NewTicker(time.Duration(r.timer.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
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

	var metrics []model.Metric
	for key, value := range r.gaugeMetric.GetStats() {
		metrics = append(metrics, model.Metric{
			ID:    key,
			Type:  r.gaugeMetric.GetName(),
			Value: &value,
		})
	}

	for key, value := range r.counterMetric.GetStats() {
		metrics = append(metrics, model.Metric{
			ID:    key,
			Type:  r.counterMetric.GetName(),
			Delta: &value,
		})
	}

	if len(metrics) > 0 {
		r.jobs <- metrics
		err := <-r.results
		if err != nil {
			logger.Log.Info(err.Error())
		}
	}

	logger.Log.Info("Finish reporting metrics...")
}

func (r *Reporter) worker(ctx context.Context, id int, jobs <-chan []model.Metric, results chan<- error) {
	for {
		select {
		case job := <-jobs:
			logger.Log.Info(fmt.Sprintf("The worker %d started the task", id), logger.Field{Key: "Metrics", Value: job})
			results <- r.updateMetrics(ctx, job...)
			logger.Log.Info(fmt.Sprintf("The worker %d ended the task", id), logger.Field{Key: "Metrics", Value: job})
		case <-ctx.Done():
			return
		}
	}
}

func (r *Reporter) updateMetrics(ctx context.Context, metrics ...model.Metric) error {
	return retry.Retry(
		func() error {
			return r.client.UpdateMetric(ctx, metrics...)
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
