package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/a-romash/grpc-calculator/sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	if err = a.gRPCServer.Serve(l); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.gRPCServer.GracefulStop()
}
