package storage_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/c2pc/go-musthave-metrics/internal/database"
	"github.com/stretchr/testify/assert"

	"github.com/c2pc/go-musthave-metrics/internal/storage"
)

func TestGaugeStorage_GetName(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	if gaugeStorage.GetName() != "gauge" {
		t.Error("Gauge storage name not set properly")
	}
}

func TestGaugeStorage_Set_Memory(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
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
			err := gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
			assert.NoError(t, err)

			got, err := gaugeStorage.Get(context.Background(), tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestGaugeStorage_Set_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   float64
		err     error
		mockgen func()
	}{
		{
			name:  "Error -> Begin",
			key:   "key1",
			value: 10,
			err:   errors.New("begin some error"),
			mockgen: func() {
				mock.ExpectBegin().WillReturnError(errors.New("begin some error"))
			},
		},
		{
			name:  "Error -> Rollback",
			key:   "key1",
			value: 10,
			err:   errors.New("rollback some error"),
			mockgen: func() {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(errors.New("rollback some error"))
			},
		},
		{
			name:  "Error -> Commit",
			key:   "key1",
			value: 10,
			err:   errors.New("commit some error"),
			mockgen: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO gauges (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key1", float64(10)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit some error"))
			},
		},
		{
			name:  "Error",
			key:   "key1",
			value: 10,
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO gauges (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key1", float64(10)).
					WillReturnError(errors.New("some error"))
				mock.ExpectRollback().WillReturnError(nil)
			},
		},
		{
			name:  "Success",
			key:   "key2",
			value: 11,
			err:   nil,
			mockgen: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO gauges (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key2", float64(11)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			err = gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tt.err, err.Error())
			}
		})
	}
}

func TestGaugeStorage_SetString(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
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
			err := gaugeStorage.SetString(context.Background(), storage.Value[string]{Key: tt.key, Value: tt.value})
			if tt.expectedError {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.Get(context.Background(), tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestGaugeStorage_GetKey_Memory(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.Get(context.Background(), tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestGaugeStorage_GetKey_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   float64
		err     error
		mockgen func()
	}{
		{
			name:  "Error",
			key:   "key1",
			value: 10,
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM gauges WHERE (.+) LIMIT 1$").
					WithArgs("key1").
					WillReturnError(errors.New("some error"))
			},
		},
		{
			name:  "Success",
			key:   "key2",
			value: 11,
			err:   nil,
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM gauges WHERE (.+) LIMIT 1$").
					WithArgs("key2").
					WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(11))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			value, err := gaugeStorage.Get(context.Background(), tt.key)
			if tt.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, value)
			} else {
				assert.EqualError(t, tt.err, err.Error())
			}
		})
	}
}

func TestGaugeStorage_GetString(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetString(context.Background(), tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestGaugeStorage_GetAll_Memory(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetAll(context.Background())
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}

func TestGaugeStorage_GetAll_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   map[string]float64
		err     error
		mockgen func()
	}{
		{
			name:  "Error",
			key:   "key1",
			value: nil,
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM gauges$").
					WillReturnError(errors.New("some error"))
			},
		},
		{
			name:  "Success",
			key:   "key2",
			value: map[string]float64{"key1": 10, "key2": 20},
			err:   nil,
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM gauges$").
					WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("key1", 10).AddRow("key2", 20))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			value, err := gaugeStorage.GetAll(context.Background())
			if tt.err == nil {
				assert.NoError(t, err)
				eq := reflect.DeepEqual(tt.value, value)
				if !eq {
					t.Errorf("got %v, want %v", value, tt.value)
				}
			} else {
				assert.EqualError(t, tt.err, err.Error())
			}
		})
	}
}

func TestGaugeStorage_GetAllString(t *testing.T) {
	gaugeStorage, err := storage.NewGaugeStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := gaugeStorage.Set(context.Background(), storage.Value[float64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := gaugeStorage.GetAllString(context.Background())
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("PollStats() = %v, want %v", got, tt.got)
			}
		})
	}
}
