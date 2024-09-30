package metric_test

import (
	"testing"

	"github.com/c2pc/go-musthave-metrics/internal/metric"
)

var gaugeMetricKeys = []string{
	metric.GaugeAllocKey,
	metric.GaugeBuckHashSysKey,
	metric.GaugeFreesKey,
	metric.GaugeGCCPUFractionKey,
	metric.GaugeGCSysKey,
	metric.GaugeHeapAllocKey,
	metric.GaugeHeapIdleKey,
	metric.GaugeHeapInuseKey,
	metric.GaugeHeapObjectsKey,
	metric.GaugeHeapReleasedKey,
	metric.GaugeHeapSysKey,
	metric.GaugeLastGCKey,
	metric.GaugeLookupsKey,
	metric.GaugeMCacheInuseKey,
	metric.GaugeMCacheSysKey,
	metric.GaugeMSpanInuseKey,
	metric.GaugeMSpanSysKey,
	metric.GaugeMallocsKey,
	metric.GaugeNextGCKey,
	metric.GaugeNumForcedGCKey,
	metric.GaugeNumGCKey,
	metric.GaugeOtherSysKey,
	metric.GaugePauseTotalNsKey,
	metric.GaugeStackInuseKey,
	metric.GaugeStackSysKey,
	metric.GaugeSysKey,
	metric.GaugeTotalAllocKey,
	metric.GaugeRandomValueKey,
}

func TestGaugeMetric_GetName(t *testing.T) {
	gaugeMetric := metric.NewGaugeMetric()

	if gaugeMetric.GetName() != "gauge" {
		t.Error("Gauge metric name not set properly")
	}
}

func TestGaugeMetric_PollStats(t *testing.T) {
	gaugeMetric := metric.NewGaugeMetric()

	gaugeMetric.PollStats()
	stats := gaugeMetric.GetStats()

	for _, key := range gaugeMetricKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("PollStats() = %+v, want %+v", stats, gaugeMetricKeys)
		}
	}
}

func TestGaugeMetric_GetStats(t *testing.T) {
	gaugeMetric := metric.NewGaugeMetric()

	gaugeMetric.PollStats()
	stats := gaugeMetric.GetStats()

	for _, key := range gaugeMetricKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("PollStats() = %+v, want %+v", stats, gaugeMetricKeys)
		}
	}
}
