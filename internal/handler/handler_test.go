package handler_test

import (
	"fmt"
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		gaugeStorage   storage.Storage[float64]
		counterStorage storage.Storage[int64]
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty storages",
			args: args{
				nil, nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "empty gauge storage",
			args: args{
				nil, storage.NewCounterStorage(),
			},
			wantErr: assert.Error,
		},
		{
			name: "empty counter storage",
			args: args{
				storage.NewGaugeStorage(), nil,
			},
			wantErr: assert.Error,
		},
		{
			name: "success",
			args: args{
				storage.NewGaugeStorage(), storage.NewCounterStorage(),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.NewHandler(tt.args.gaugeStorage, tt.args.counterStorage)
			if !tt.wantErr(t, err, fmt.Sprintf("NewHandler(%v, %v)", tt.args.gaugeStorage, tt.args.counterStorage)) {
				return
			}
		})
	}
}
