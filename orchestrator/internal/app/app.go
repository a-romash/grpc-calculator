package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/grpc"
	httpapp "github.com/a-romash/grpc-calculator/orchestrator/internal/app/http"
	orch "github.com/a-romash/grpc-calculator/orchestrator/internal/service/orchestrator"
)

type App struct {
	GRPCServer *grpcapp.App
	HTTPServer *httpapp.HTTPApp
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	orchService := orch.New(log, nil)

	grpcApp := grpcapp.New(log, orchService, grpcPort)
	return &App{
		GRPCServer: grpcApp,
	}
}
