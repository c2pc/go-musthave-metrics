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
	"github.com/c2pc/go-musthave-metrics/internal/worker_pool"
)

type Updater interface {
	UpdateMetric(ctx context.Context, metrics ...model.Metric) error
}

type Worker interface {
	TaskRun(task worker_pool.Task)
	TaskResult() <-chan error
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
	worker        Worker
}

func New(client Updater, workerPool Worker, timer Timer, counterMetric MetricReader[int64], gaugeMetric MetricReader[float64]) *Reporter {
	return &Reporter{
		counterMetric: counterMetric,
		gaugeMetric:   gaugeMetric,
		client:        client,
		timer:         timer,
		worker:        workerPool,
	}
}

func (r *Reporter) Run(ctx context.Context) {
	go r.poll(ctx)
	go r.report(ctx)

	<-ctx.Done()
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
		r.worker.TaskRun(func() error {
			return r.client.UpdateMetric(ctx, metrics...)
		})
		err := <-r.worker.TaskResult()
		if err != nil {
			logger.Log.Info(err.Error())
		}
	}

	logger.Log.Info("Finish reporting metrics...")
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
