package metric

import (
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/reporter"
)

const (
	CounterPollCountKey string = "PollCount"
)

type CounterMetric struct {
	mu    *sync.Mutex
	stats map[string]int64
}

func NewCounterMetric() reporter.MetricReader[int64] {
	return &CounterMetric{
		mu:    &sync.Mutex{},
		stats: make(map[string]int64),
	}
}

func (m *CounterMetric) GetName() string {
	return "counter"
}

func (m *CounterMetric) GetStats() map[string]int64 {
	return m.stats
}

func (m *CounterMetric) PollStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats[CounterPollCountKey] += 1
}
