package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/clients/sso/grpc"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/server"
	httpservice "github.com/a-romash/grpc-calculator/orchestrator/internal/service/http"
)

type HTTPApp struct {
	log    *slog.Logger
	server *server.Server
}

func New(
	log *slog.Logger,
	port int,
	addr string, // address of SSO-service
	retriesCount int,
	storage httpservice.ExpressionStorage,
	secret string,
) *HTTPApp {
	app_id, err := storage.RegisterApp(context.Background(), "Orchestrator", secret)
	if err != nil {
		log.Info("Error while registering App")
		log.Error(err.Error())
		return nil
	}

	client, err := grpc.New(context.Background(), log, addr, 5*time.Minute, retriesCount, int(app_id))
	if err != nil {
		log.Info("Error while registering grpc client")
		log.Error(err.Error())
		return nil
	}

	server := server.New(log, storage, port, secret, client)

	return &HTTPApp{
		log:    log,
		server: server,
	}
}

func (a *HTTPApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *HTTPApp) Run() error {
	const op = "httpapp.Run"

	if err := a.server.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
