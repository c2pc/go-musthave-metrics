package storage_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/c2pc/go-musthave-metrics/internal/database"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCounterStorage_GetName(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

	if counterStorage.GetName() != "counter" {
		t.Error("Counter storage name not set properly")
	}
}

func TestCounterStorage_Set_Memory(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
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
			err := counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
			assert.NoError(t, err)

			got, err := counterStorage.Get(context.Background(), tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestCounterStorage_Set_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	counterStorage, err := storage.NewCounterStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   int64
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
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectBegin()
				mock.ExpectExec("^INSERT INTO counters (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key1", 10).WillReturnError(errors.New("some error"))
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
				mock.ExpectExec("^INSERT INTO counters (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key1", 10).WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectExec("^INSERT INTO counters (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key1", 10).
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
				mock.ExpectExec("^INSERT INTO counters (.+) VALUES (.+) ON CONFLICT (.+) DO UPDATE SET (.+)$").
					WithArgs("key2", 11).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			err = counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tt.err, err.Error())
			}
		})
	}
}

func TestCounterStorage_SetString(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)
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
			err := counterStorage.SetString(context.Background(), storage.Value[string]{Key: tt.key, Value: tt.value})
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			got, err := counterStorage.Get(context.Background(), tt.key)
			assert.NoError(t, err)
			assert.Equal(t, tt.got, got)
		})
	}
}

func TestCounterStorage_GetKey_Memory(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := counterStorage.Get(context.Background(), tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestCounterStorage_GetKey_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	counterStorage, err := storage.NewCounterStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   int64
		err     error
		mockgen func()
	}{
		{
			name:  "Error",
			key:   "key1",
			value: 10,
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM counters WHERE (.+) LIMIT 1$").
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
				mock.ExpectQuery("^SELECT (.+) FROM counters WHERE (.+) LIMIT 1$").
					WithArgs("key2").
					WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(11))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			value, err := counterStorage.Get(context.Background(), tt.key)
			if tt.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.value, value)
			} else {
				assert.EqualError(t, tt.err, err.Error())
			}
		})
	}
}

func TestCounterStorage_GetString(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetString(context.Background(), tt.key)
			if !tt.set {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.got, got)
			}
		})
	}
}

func TestCounterStorage_GetAll_Memory(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetAll(context.Background())
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("got %v, want %v", got, tt.got)
			}
		})
	}
}

func TestCounterStorage_GetAll_DB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	counterStorage, err := storage.NewCounterStorage(storage.TypeDB, &database.DB{DB: mockDB})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		key     string
		value   map[string]int64
		err     error
		mockgen func()
	}{
		{
			name:  "Error",
			key:   "key1",
			value: nil,
			err:   errors.New("some error"),
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM counters$").
					WillReturnError(errors.New("some error"))
			},
		},
		{
			name:  "Success",
			key:   "key2",
			value: map[string]int64{"key1": 10, "key2": 20},
			err:   nil,
			mockgen: func() {
				mock.ExpectQuery("^SELECT (.+) FROM counters$").
					WillReturnRows(sqlmock.NewRows([]string{"key", "value"}).AddRow("key1", 10).AddRow("key2", 20))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockgen()

			value, err := counterStorage.GetAll(context.Background())
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

func TestCounterStorage_GetAllString(t *testing.T) {
	counterStorage, err := storage.NewCounterStorage(storage.TypeMemory, nil)
	assert.NoError(t, err)

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
				err := counterStorage.Set(context.Background(), storage.Value[int64]{Key: tt.key, Value: tt.value})
				assert.NoError(t, err)
			}

			got, err := counterStorage.GetAllString(context.Background())
			assert.NoError(t, err)

			eq := reflect.DeepEqual(tt.got, got)
			if !eq {
				t.Errorf("got %v, want %v", got, tt.got)
			}
		})
	}
}
