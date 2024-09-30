package storage_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/c2pc/go-musthave-metrics/internal/storage"
)

func TestCounterStorage_GetName(t *testing.T) {
	counterStorage := storage.NewCounterStorage()

	if counterStorage.GetName() != "counter" {
		t.Error("Counter storage name not set properly")
	}
}

func TestCounterStorage_Set(t *testing.T) {
	counterStorage := storage.NewCounterStorage()
	tests := []struct {
		name  string
		key   string
		value int64
		got   int64
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
			got:   25,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := counterStorage.Set(tt.key, tt.value)
			assert.NoError(t, err)

			got, err := counterStorage.Get(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestCounterStorage_SetString(t *testing.T) {
	counterStorage := storage.NewCounterStorage()
	tests := []struct {
		name          string
		key           string
		value         string
		got           int64
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
			got:           25,
			expectedError: false,
		},
		{
			name:          "Test with replace value2",
			key:           "key1",
			value:         "20",
			got:           45,
			expectedError: false,
		},
		{
			name:          "Test error to parse value",
			key:           "key1",
			value:         "invalid",
			got:           45,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := counterStorage.SetString(tt.key, tt.value)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			got, err := counterStorage.Get(tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestCounterStorage_GetKey(t *testing.T) {
	counterStorage := storage.NewCounterStorage()

	tests := []struct {
		name  string
		key   string
		value int64
		got   int64
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
			got:   25,
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   45,
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := counterStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := counterStorage.Get(tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestCounterStorage_GetString(t *testing.T) {
	counterStorage := storage.NewCounterStorage()

	tests := []struct {
		name  string
		key   string
		value int64
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
			got:   "25",
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   "45",
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := counterStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetString(tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestCounterStorage_GetAll(t *testing.T) {
	counterStorage := storage.NewCounterStorage()

	tests := []struct {
		name  string
		key   string
		value int64
		got   map[string]int64
		set   bool
	}{
		{
			name:  "Test with existing key",
			key:   "key1",
			value: 10,
			got:   map[string]int64{"key1": 10},
			set:   true,
		},
		{
			name:  "Test with non-existing key",
			key:   "key2",
			value: 0,
			got:   map[string]int64{"key1": 10},
			set:   false,
		},
		{
			name:  "Test with replace value",
			key:   "key1",
			value: 15,
			got:   map[string]int64{"key1": 25},
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   map[string]int64{"key1": 45},
			set:   true,
		},
		{
			name:  "Test with existing key2",
			key:   "key2",
			value: 1,
			got:   map[string]int64{"key1": 45, "key2": 1},
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := counterStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetAll()
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}

func TestCounterStorage_GetAllString(t *testing.T) {
	counterStorage := storage.NewCounterStorage()

	tests := []struct {
		name  string
		key   string
		value int64
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
			got:   map[string]string{"key1": "25"},
			set:   true,
		},
		{
			name:  "Test with replace value2",
			key:   "key1",
			value: 20,
			got:   map[string]string{"key1": "45"},
			set:   true,
		},
		{
			name:  "Test with existing key2",
			key:   "key2",
			value: 1,
			got:   map[string]string{"key1": "45", "key2": "1"},
			set:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				err := counterStorage.Set(tt.key, tt.value)
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetAllString()
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}
