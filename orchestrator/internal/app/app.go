package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/grpc"
	httpapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/http"
	orch "github.com/a-romash/grpc-calculator/orchestrator/internal/service/orchestrator"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.HTTPApp
}

func New(log *slog.Logger, grpcPort int, storage *postgres.Postgresql, tokenTTL time.Duration, httpPort int, addr string, retriesCount int, secret string) *App {
	orchService := orch.New(log, storage)

	grpcApp := grpcapp.New(log, orchService, grpcPort)
	httpApp := httpapp.New(log, httpPort, addr, retriesCount, storage, secret)

	return &App{
		GRPCServer: grpcApp,
		HTTPServer: httpApp,
	}
}
