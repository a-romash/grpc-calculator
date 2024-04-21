package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/clients/sso/grpc"
	middleware "github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/middlewares"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/routes"
	"github.com/gorilla/mux"
)

type HTTPApp struct {
	log      *slog.Logger
	serveMux *mux.Router
	port     int
	client   *grpc.Client
}

func New(
	log *slog.Logger,
	port int,
	addr string, // address of SSO-services
	retriesCount int,
	app_id int,
) *HTTPApp {
	serveMux := routes.New()

	client, err := grpc.New(context.Background(), log, addr, 5*time.Minute, retriesCount, app_id)
	if err != nil {
		return nil
	}

	return &HTTPApp{
		log:      log,
		serveMux: serveMux,
		port:     port,
		client:   client,
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
		return err
	}
	return nil
}
