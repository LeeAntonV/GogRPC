package main

import (
	"fmt"
	"gRPC/internal/app"
	"gRPC/internal/config"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)
	log.Info("starting application", slog.Any("cfg", cfg))

	application := app.New(log, cfg.GRPC.Port, cfg.TokenTTL)

	go func() {
		application.GRPCServer.MustRun()
	}()

	//TODO: Initialize app

	//TODO: start GRPC-server

	//Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	application.GRPCServer.Stop()

	log.Info("Application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
