package metric

import (
	"github.com/c2pc/go-musthave-metrics/internal/reporter"
	"math/rand"
	"runtime"
	"sync"
)

const (
	GaugeAllocKey         string = "Alloc"
	GaugeBuckHashSysKey   string = "BuckHashSys"
	GaugeFreesKey         string = "Frees"
	GaugeGCCPUFractionKey string = "GCCPUFraction"
	GaugeGCSysKey         string = "GCSys"
	GaugeHeapAllocKey     string = "HeapAlloc"
	GaugeHeapIdleKey      string = "HeapIdle"
	GaugeHeapInuseKey     string = "HeapInuse"
	GaugeHeapObjectsKey   string = "HeapObjects"
	GaugeHeapReleasedKey  string = "HeapReleased"
	GaugeHeapSysKey       string = "HeapSys"
	GaugeLastGCKey        string = "LastGC"
	GaugeLookupsKey       string = "Lookups"
	GaugeMCacheInuseKey   string = "MCacheInuse"
	GaugeMCacheSysKey     string = "MCacheSys"
	GaugeMSpanInuseKey    string = "MSpanInuse"
	GaugeMSpanSysKey      string = "MSpanSys"
	GaugeMallocsKey       string = "Mallocs"
	GaugeNextGCKey        string = "NextGC"
	GaugeNumForcedGCKey   string = "NumForcedGC"
	GaugeNumGCKey         string = "NumGC"
	GaugeOtherSysKey      string = "OtherSys"
	GaugePauseTotalNsKey  string = "PauseTotalNs"
	GaugeStackInuseKey    string = "StackInuse"
	GaugeStackSysKey      string = "StackSys"
	GaugeSysKey           string = "Sys"
	GaugeTotalAllocKey    string = "TotalAlloc"
	GaugeRandomValueKey   string = "RandomValue"
)

type GaugeMetric struct {
	mu    *sync.Mutex
	stats map[string]float64
}

func NewGaugeMetric() reporter.MetricReader[float64] {
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

	m.stats[GaugeAllocKey] = float64(memStats.Alloc)
	m.stats[GaugeBuckHashSysKey] = float64(memStats.BuckHashSys)
	m.stats[GaugeFreesKey] = float64(memStats.Frees)
	m.stats[GaugeGCCPUFractionKey] = memStats.GCCPUFraction
	m.stats[GaugeGCSysKey] = float64(memStats.GCSys)
	m.stats[GaugeHeapAllocKey] = float64(memStats.HeapAlloc)
	m.stats[GaugeHeapIdleKey] = float64(memStats.HeapIdle)
	m.stats[GaugeHeapInuseKey] = float64(memStats.HeapInuse)
	m.stats[GaugeHeapObjectsKey] = float64(memStats.HeapObjects)
	m.stats[GaugeHeapReleasedKey] = float64(memStats.HeapReleased)
	m.stats[GaugeHeapSysKey] = float64(memStats.HeapSys)
	m.stats[GaugeLastGCKey] = float64(memStats.LastGC)
	m.stats[GaugeLookupsKey] = float64(memStats.Lookups)
	m.stats[GaugeMCacheInuseKey] = float64(memStats.MCacheInuse)
	m.stats[GaugeMCacheSysKey] = float64(memStats.MCacheSys)
	m.stats[GaugeMSpanInuseKey] = float64(memStats.MSpanInuse)
	m.stats[GaugeMSpanSysKey] = float64(memStats.MSpanSys)
	m.stats[GaugeMallocsKey] = float64(memStats.Mallocs)
	m.stats[GaugeNextGCKey] = float64(memStats.NextGC)
	m.stats[GaugeNumForcedGCKey] = float64(memStats.NumForcedGC)
	m.stats[GaugeNumGCKey] = float64(memStats.NumGC)
	m.stats[GaugeOtherSysKey] = float64(memStats.OtherSys)
	m.stats[GaugePauseTotalNsKey] = float64(memStats.PauseTotalNs)
	m.stats[GaugeStackInuseKey] = float64(memStats.StackInuse)
	m.stats[GaugeStackSysKey] = float64(memStats.StackSys)
	m.stats[GaugeSysKey] = float64(memStats.Sys)
	m.stats[GaugeTotalAllocKey] = float64(memStats.TotalAlloc)
	m.stats[GaugeRandomValueKey] = rand.Float64()
}
