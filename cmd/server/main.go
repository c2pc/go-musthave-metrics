package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/c2pc/go-musthave-metrics/cmd/server/config"
	"github.com/c2pc/go-musthave-metrics/internal/handler"
	"github.com/c2pc/go-musthave-metrics/internal/server"
	"github.com/c2pc/go-musthave-metrics/internal/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
		return
	}

	gaugeStorage := storage.NewGaugeStorage()
	counterStorage := storage.NewCounterStorage()

	handlers, err := handler.NewHandler(gaugeStorage, counterStorage)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := server.NewServer(handlers, cfg.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		fmt.Printf("Starting Server on %s\n", cfg.Address)
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
