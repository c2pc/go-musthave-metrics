package metric

import (
	"math/rand"
	"runtime"
	"sync"
)

const (
	GaugeAllocKey         Key = "Alloc"
	GaugeBuckHashSysKey   Key = "BuckHashSys"
	GaugeFreesKey         Key = "Frees"
	GaugeGCCPUFractionKey Key = "GCCPUFraction"
	GaugeGCSysKey         Key = "GCSys"
	GaugeHeapAllocKey     Key = "HeapAlloc"
	GaugeHeapIdleKey      Key = "HeapIdle"
	GaugeHeapInuseKey     Key = "HeapInuse"
	GaugeHeapObjectsKey   Key = "HeapObjects"
	GaugeHeapReleasedKey  Key = "HeapReleased"
	GaugeHeapSysKey       Key = "HeapSys"
	GaugeLastGCKey        Key = "LastGC"
	GaugeLookupsKey       Key = "Lookups"
	GaugeMCacheInuseKey   Key = "MCacheInuse"
	GaugeMCacheSysKey     Key = "MCacheSys"
	GaugeMSpanInuseKey    Key = "MSpanInuse"
	GaugeMSpanSysKey      Key = "MSpanSys"
	GaugeMallocsKey       Key = "Mallocs"
	GaugeNextGCKey        Key = "NextGC"
	GaugeNumForcedGCKey   Key = "NumForcedGC"
	GaugeNumGCKey         Key = "NumGC"
	GaugeOtherSysKey      Key = "OtherSys"
	GaugePauseTotalNsKey  Key = "PauseTotalNs"
	GaugeStackInuseKey    Key = "StackInuse"
	GaugeStackSysKey      Key = "StackSys"
	GaugeSysKey           Key = "Sys"
	GaugeTotalAllocKey    Key = "TotalAlloc"
	GaugeRandomValueKey   Key = "RandomValue"
)

type GaugeMetric struct {
	mu    *sync.Mutex
	stats map[Key]float64
}

func NewGaugeMetric() *GaugeMetric {
	return &GaugeMetric{
		mu:    &sync.Mutex{},
		stats: make(map[Key]float64),
	}
}

func (m *GaugeMetric) GetName() string {
	return "gauge"
}

func (m *GaugeMetric) GetStats() map[Key]float64 {
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
