package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/app"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/config"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/lib/logger/handlers/slogpretty"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/storage/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	storage, err := postgres.Connect(cfg.DatabaseUrl)
	if err != nil {
		panic(err)
	}

	application := app.New(log, cfg.GRPC.Port, storage, cfg.TokenTTL, cfg.HTTP.Port, cfg.GRPCClient.Addr, cfg.GRPCClient.RetriesCount, cfg.Secret)
	if application.HTTPServer == nil {
		panic("httpserver is nil!!1!")
	}
	go func() {
		application.HTTPServer.MustRun()
	}()
	go func() {
		application.GRPCServer.MustRun()
	}()

	// Graceful stop

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	application.GRPCServer.Stop()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettyLogger()
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

func setupPrettyLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
