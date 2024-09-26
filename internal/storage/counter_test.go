package storage_test

import (
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
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
