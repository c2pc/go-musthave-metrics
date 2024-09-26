package storage_test

import (
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGaugeStorage_GetName(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()

	if gaugeStorage.GetName() != "gauge" {
		t.Error("Gauge storage name not set properly")
	}
}

func TestGaugeStorage_Set(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	tests := []struct {
		name  string
		key   string
		value float64
		got   float64
	}{
		{
			name:  "Test with non-existing key",
			key:   "key1",
			value: 10,
			got:   10,
		},
		{
			name:  "Test with non-existing key2",
			key:   "key2",
			value: 0,
			got:   0,
		},
		{
			name:  "Test with replace value",
			key:   "key1",
			value: 15,
			got:   15,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gaugeStorage.Set(tt.key, tt.value)
			assert.NoError(t, err)

			got, err := gaugeStorage.Get(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestGaugeStorage_GetKey(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()

	tests := []struct {
		name  string
		key   string
		value float64
		got   float64
		set   bool
	}{
		{
			name:  "Test with existing key",
			key:   "key1",
			value: 10,
			got:   10,
			set:   true,
		},
		{
			name:  "Test with non-existing key",
			key:   "key2",
			value: 0,
			set:   false,
		},
		{
			name:  "Test with replace value",
			key:   "key1",
			value: 15,
			got:   15,
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   20,
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := gaugeStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.Get(tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}
