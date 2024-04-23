package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/clients/sso/grpc"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/handlers"
	middleware "github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/middlewares"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/routes"
	"github.com/gorilla/mux"
)

type HTTPApp struct {
	log        *slog.Logger
	serveMux   *mux.Router
	port       int
	sso_client *grpc.Client
}

func New(
	log *slog.Logger,
	port int,
	addr string, // address of SSO-service
	retriesCount int,
	storage handlers.ExpressionStorage,
	secret string,
) *HTTPApp {
	serveMux := routes.New(log, storage)

	app_id, err := storage.RegisterApp(context.Background(), "Orchestrator", secret)
	if err != nil {
		return nil
	}

	client, err := grpc.New(context.Background(), log, addr, 5*time.Minute, retriesCount, int(app_id))
	if err != nil {
		return nil
	}

	return &HTTPApp{
		log:        log,
		serveMux:   serveMux,
		port:       port,
		sso_client: client,
	}
}

func (a *HTTPApp) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *HTTPApp) Run() error {
	const op = "httpapp.Run"

	// Добавляем middleware для перехвата паник
	muxWithPanicHandler := middleware.PanicMiddleware(a.serveMux)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: muxWithPanicHandler,
	}

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
