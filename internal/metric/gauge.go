package metric

import (
	"math/rand"
	"runtime"
	"sync"
)

type GaugeMetric struct {
	mu    *sync.Mutex
	stats map[string]float64
}

func NewGaugeMetric() *GaugeMetric {
	return &GaugeMetric{
		mu:    &sync.Mutex{},
		stats: make(map[string]float64),
	}
}

func (m *GaugeMetric) GetName() string {
	return "gauge"
}

func (m *GaugeMetric) GetStats() map[string]float64 {
	return m.stats
}

func (m *GaugeMetric) PollStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats["Alloc"] = float64(memStats.Alloc)
	m.stats["BuckHashSys"] = float64(memStats.BuckHashSys)
	m.stats["Frees"] = float64(memStats.Frees)
	m.stats["GCCPUFraction"] = memStats.GCCPUFraction
	m.stats["GCSys"] = float64(memStats.GCSys)
	m.stats["HeapAlloc"] = float64(memStats.HeapAlloc)
	m.stats["HeapIdle"] = float64(memStats.HeapIdle)
	m.stats["HeapInuse"] = float64(memStats.HeapInuse)
	m.stats["HeapObjects"] = float64(memStats.HeapObjects)
	m.stats["HeapReleased"] = float64(memStats.HeapReleased)
	m.stats["HeapSys"] = float64(memStats.HeapSys)
	m.stats["LastGC"] = float64(memStats.LastGC)
	m.stats["Lookups"] = float64(memStats.Lookups)
	m.stats["MCacheInuse"] = float64(memStats.MCacheInuse)
	m.stats["MCacheSys"] = float64(memStats.MCacheSys)
	m.stats["MSpanInuse"] = float64(memStats.MSpanInuse)
	m.stats["MSpanSys"] = float64(memStats.MSpanSys)
	m.stats["Mallocs"] = float64(memStats.Mallocs)
	m.stats["NextGC"] = float64(memStats.NextGC)
	m.stats["NumForcedGC"] = float64(memStats.NumForcedGC)
	m.stats["NumGC"] = float64(memStats.NumGC)
	m.stats["OtherSys"] = float64(memStats.OtherSys)
	m.stats["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	m.stats["StackInuse"] = float64(memStats.StackInuse)
	m.stats["StackSys"] = float64(memStats.StackSys)
	m.stats["Sys"] = float64(memStats.Sys)
	m.stats["TotalAlloc"] = float64(memStats.TotalAlloc)
	m.stats["RandomValue"] = rand.Float64()

	return
}
