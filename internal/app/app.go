package app

import (
	grpcapp "gRPC/internal/app/grpc"
	"gRPC/internal/services/auth"
	"gRPC/internal/storage/postgres"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, port int, storagePath string, tokenTTL time.Duration) *App {
	storage, err := postgres.New(storagePath)
	if err != nil {
		panic(err)
	}

	authservice := auth.New(
		log,
		storage,
		storage,
		storage,
		storage,
		tokenTTL,
	)
	grpcApp := grpcapp.New(log, authservice, port)

	return &App{
		GRPCServer: grpcApp,
	}
}
