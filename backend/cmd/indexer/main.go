// Package main is the entry point for the VIGILUM indexer service.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting VIGILUM Indexer", "version", "0.1.0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// TODO: Connect to blockchain RPCs
	// TODO: Initialize block processors
	// TODO: Start mempool listener
	// TODO: Start indexing pipeline

	<-ctx.Done()
	slog.Info("VIGILUM Indexer shutdown complete")
}
