package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	orchgrpc "github.com/a-romash/grpc-calculator/orchestrator/internal/grpc/orchestrator"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, orchService orchgrpc.Orchestrator, port int) *App {
	gRPCServer := grpc.NewServer()

	orchgrpc.Register(gRPCServer, orchService)

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
