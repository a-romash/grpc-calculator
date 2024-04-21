package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/a-romash/grpc-calculator/sso/internal/app/grpc"
	"github.com/a-romash/grpc-calculator/sso/internal/service/auth"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	authService := auth.New(log, nil, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)
	return &App{
		GRPCServer: grpcApp,
	}
}
