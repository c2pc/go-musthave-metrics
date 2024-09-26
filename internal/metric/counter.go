package metric

import (
	"sync"
)

const (
	CounterPollCountKey Key = "PollCount"
)

type CounterMetric struct {
	mu    *sync.Mutex
	stats map[Key]int64
}

func NewCounterMetric() *CounterMetric {
	return &CounterMetric{
		mu:    &sync.Mutex{},
		stats: make(map[Key]int64),
	}
}

func (m *CounterMetric) GetName() string {
	return "counter"
}

func (m *CounterMetric) GetStats() map[Key]int64 {
	return m.stats
}

func (m *CounterMetric) PollStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats[CounterPollCountKey] += 1

	return
}
