package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-romash/grpc-calculator/agent/internal/app"
	"github.com/a-romash/grpc-calculator/agent/internal/config"
	"github.com/a-romash/grpc-calculator/agent/internal/lib/logger/handlers/slogpretty"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	application := app.New(log, cfg.GRPCClient.Addr, cfg.GRPCClient.RetriesCount, cfg.CountCalcs, map[string]time.Duration{
		"+": cfg.Durations.Plus,
		"-": cfg.Durations.Minus,
		"*": cfg.Durations.Mult,
		"/": cfg.Durations.Del,
		"^": cfg.Durations.Pow,
	})

	application.GRPCClient.MustRun()

	// Graceful stop

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-stop

	application.GRPCClient.Stop()
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
