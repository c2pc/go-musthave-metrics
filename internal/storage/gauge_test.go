package storage_test

import (
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"reflect"
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

func TestGaugeStorage_SetString(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()
	tests := []struct {
		name          string
		key           string
		value         string
		got           float64
		expectedError bool
	}{
		{
			name:          "Test with non-existing key",
			key:           "key1",
			value:         "10",
			got:           10,
			expectedError: false,
		},
		{
			name:          "Test with non-existing key2",
			key:           "key2",
			value:         "0",
			got:           0,
			expectedError: false,
		},
		{
			name:          "Test with replace value",
			key:           "key1",
			value:         "15",
			got:           15,
			expectedError: false,
		},
		{
			name:          "Test with replace value2",
			key:           "key1",
			value:         "20",
			got:           20,
			expectedError: false,
		},
		{
			name:          "Test error to parse value",
			key:           "key3",
			value:         "invalid",
			got:           0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gaugeStorage.SetString(tt.key, tt.value)
			if tt.expectedError {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

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

func TestGaugeStorage_GetString(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()

	tests := []struct {
		name  string
		key   string
		value float64
		got   string
		set   bool
	}{
		{
			name:  "Test with existing key",
			key:   "key1",
			value: 10,
			got:   "10",
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
			got:   "15",
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   "20",
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := gaugeStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetString(tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestGaugeStorage_GetAll(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()

	tests := []struct {
		name  string
		key   string
		value float64
		got   map[string]float64
		set   bool
	}{
		{
			name:  "Test with existing key",
			key:   "key1",
			value: 10,
			got:   map[string]float64{"key1": 10},
			set:   true,
		},
		{
			name:  "Test with non-existing key",
			key:   "key2",
			value: 0,
			got:   map[string]float64{"key1": 10},
			set:   false,
		},
		{
			name:  "Test with replace value",
			key:   "key1",
			value: 15,
			got:   map[string]float64{"key1": 15},
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   map[string]float64{"key1": 20},
			set:   true,
		},
		{
			name:  "Test with existing key2",
			key:   "key2",
			value: 1,
			got:   map[string]float64{"key1": 20, "key2": 1},
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := gaugeStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetAll()
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}

func TestGaugeStorage_GetAllString(t *testing.T) {
	gaugeStorage := storage.NewGaugeStorage()

	tests := []struct {
		name  string
		key   string
		value float64
		got   map[string]string
		set   bool
	}{
		{
			name:  "Test with existing key",
			key:   "key1",
			value: 10,
			got:   map[string]string{"key1": "10"},
			set:   true,
		},
		{
			name:  "Test with non-existing key",
			key:   "key2",
			value: 0,
			got:   map[string]string{"key1": "10"},
			set:   false,
		},
		{
			name:  "Test with replace value",
			key:   "key1",
			value: 15,
			got:   map[string]string{"key1": "15"},
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   map[string]string{"key1": "20"},
			set:   true,
		},
		{
			name:  "Test with existing key2",
			key:   "key2",
			value: 1,
			got:   map[string]string{"key1": "20", "key2": "1"},
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := gaugeStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetAllString()
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}
