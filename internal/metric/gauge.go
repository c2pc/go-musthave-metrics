package metric

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/c2pc/go-musthave-metrics/internal/reporter"
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

const (
	TotalMemory    string = "TotalMemory"
	FreeMemory     string = "FreeMemory"
	CPUUtilization string = "CPUutilization"
)

type GaugeMetric struct {
	mu    sync.RWMutex
	stats map[string]float64
}

func NewGaugeMetric() reporter.MetricReader[float64] {
	return &GaugeMetric{
		mu:    sync.RWMutex{},
		stats: make(map[string]float64),
	}
}

func (m *GaugeMetric) GetName() string {
	return "gauge"
}

func (m *GaugeMetric) GetStats() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

func (m *GaugeMetric) PollStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		logger.Log.Info("Read Mem Stats", logger.Field{Key: "Stats", Value: memStats})

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
	}()

	go func() {
		defer wg.Done()
		cpustat, err := cpu.Percent(0, false)
		if err != nil {
			logger.Log.Info("Error to read CPU Stats", logger.Error(err))
			return
		}

		logger.Log.Info("Read CPU Stats", logger.Field{Key: "Stats", Value: cpustat})

		virtualMemory, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Info("Error to read Virtual Memory Stats", logger.Error(err))
			return
		}

		logger.Log.Info("Read Virtual Memory Stats", logger.Field{Key: "Stats", Value: cpustat})

		m.stats[TotalMemory] = float64(virtualMemory.Total)
		m.stats[FreeMemory] = float64(virtualMemory.Free)

		for i := 0; i < len(cpustat); i++ {
			m.stats[CPUUtilization+strconv.Itoa(i)] = cpustat[i]
		}
	}()

	wg.Wait()
}
