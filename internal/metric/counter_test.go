package metric_test

import (
	"github.com/c2pc/go-musthave-metrics/internal/metric"
	"reflect"
	"testing"
)

func TestCounterMetric_GetName(t *testing.T) {
	counterMetric := metric.NewCounterMetric()

	if counterMetric.GetName() != "counter" {
		t.Error("Counter metric name not set properly")
	}
}

func TestCounterMetric_PollStats(t *testing.T) {
	counterMetric := metric.NewCounterMetric()

	tests := []struct {
		name string
		want map[metric.Key]int64
	}{
		{
			name: "first polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 1},
		},
		{
			name: "second polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 2},
		},
		{
			name: "third polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 3},
		},
		{
			name: "fourth polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterMetric.PollStats()
			stats := counterMetric.GetStats()
			if len(stats) != len(tt.want) {
				t.Errorf("PollStats() = %v, want %v", stats, tt.want)
			}

			eq := reflect.DeepEqual(tt.want, stats)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", stats, tt.want)
			}
		})
	}
}

func TestCounterMetric_GetStats(t *testing.T) {
	counterMetric := metric.NewCounterMetric()

	tests := []struct {
		name string
		want map[metric.Key]int64
		run  bool
	}{
		{
			name: "first polling",
			want: map[metric.Key]int64{},
			run:  false,
		},
		{
			name: "second polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 1},
			run:  true,
		},
		{
			name: "third polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 2},
			run:  true,
		},
		{
			name: "fourth polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 2},
			run:  false,
		},
		{
			name: "fifth polling",
			want: map[metric.Key]int64{metric.CounterPollCountKey: 3},
			run:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.run {
				counterMetric.PollStats()
			}

			stats := counterMetric.GetStats()
			if len(stats) != len(tt.want) {
				t.Errorf("PollStats() = %v, want %v", stats, tt.want)
			}

			eq := reflect.DeepEqual(tt.want, stats)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", stats, tt.want)
			}
		})
	}
}
