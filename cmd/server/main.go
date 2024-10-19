package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/c2pc/go-musthave-metrics/cmd/server/config"
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/logger"
	"github.com/c2pc/go-musthave-metrics/internal/server"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"go.uber.org/zap"
)

func main() {
	err := logger.Initialize("info")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v\n", err)
	}
	defer logger.Log.Sync()

	cfg, err := config.Parse()
	if err != nil {
		logger.Log.Fatal("failed to parse config", zap.Error(err))
		return
	}

	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handlers, err := handler.NewHandler(gaugeStorage, counterStorage)
	if err != nil {
		logger.Log.Fatal("failed to init handlers", zap.Error(err))
	}

	httpServer := server.NewServer(handlers, cfg.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		logger.Log.Info("Starting Server", zap.String("address", cfg.Address))
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Info("Error to ListenAndServe", zap.Error(err))
			close(quit)
		}
	}()

	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.Stop(ctx); err != nil {
		logger.Log.Info("Failed to Stop Server", zap.Error(err))
	}

	logger.Log.Info("Shutting Down Server")
}
