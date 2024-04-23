package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/grpc"
	httpapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/http"
	orch "github.com/a-romash/grpc-calculator/orchestrator/internal/service/orchestrator"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.HTTPApp
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration, httpPort int, addr string, retriesCount int, secret string) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	orchService := orch.New(log, storage)

	grpcApp := grpcapp.New(log, orchService, grpcPort)
	httpApp := httpapp.New(log, httpPort, addr, retriesCount, storage, secret)

	return &App{
		GRPCServer: grpcApp,
		HTTPServer: httpApp,
	}
}

func (a *App) Run() {
	go func() {
		a.GRPCServer.Run()
	}()

	go func() {
		a.HTTPServer.Run()
	}()
}
