package metric

import (
	"sync"
)

type CounterMetric struct {
	mu    *sync.Mutex
	stats map[string]int64
}

func NewCounterMetric() *CounterMetric {
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

	m.stats["PollCount"] += 1

	return
}
