package app

import (
	"log/slog"
	"time"

	grpcc "github.com/a-romash/grpc-calculator/agent/internal/app/grpc_client"
)

type App struct {
	GRPCClient *grpcc.GRPCCApp
}

func New(
	log *slog.Logger,
	addr string,
	retriesCount int,
	countCalcs int,
	durations map[string]time.Duration,
) *App {
	grpccApp := grpcc.New(log, addr, retriesCount, countCalcs, durations)

	return &App{
		GRPCClient: grpccApp,
	}
}
