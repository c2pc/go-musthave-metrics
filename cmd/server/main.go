package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/server"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultPort = "8080"
)

func main() {
	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handlers := handler.NewHandler(gaugeStorage, counterStorage)

	httpServer := server.NewServer(handlers, defaultPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Printf("Starting Server on port %s\n", defaultPort)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Error to ListenAndServe err: %s\n", err.Error())
			close(quit)
		}
	}()

	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.Stop(ctx); err != nil {
		fmt.Printf("Failed to Stop Server: %v\n", err)
	}

	fmt.Println("Shutting Down Server")
}
