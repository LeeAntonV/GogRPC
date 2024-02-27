package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type App struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       int
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func New(log *slog.Logger, port int) *App {
	grpcServer := grpc.NewServer()
	return &App{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "App is running"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Server is running", slog.String("addr", l.Addr().String()))

	if err = a.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "App is not running anymore"

	a.log.With(
		slog.String("op", op)).Info("Server is stopped", slog.Int("port", a.port))

	a.grpcServer.GracefulStop()
}
