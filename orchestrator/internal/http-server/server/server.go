package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/clients/sso/grpc"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	middleware "github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/middlewares"
	httpservice "github.com/a-romash/grpc-calculator/orchestrator/internal/service/http"
	"github.com/gorilla/mux"
)

type Server struct {
	log         *slog.Logger
	storage     httpservice.ExpressionStorage
	secret      string
	server      *http.Server
	httpService HttpService
}

type HttpService interface {
	GetExpressionById(
		ctx context.Context,
		id string,
		uid int,
	) (*models.Expression, error)
	EvaluateExpression(
		ctx context.Context,
		expression *models.Expression,
		uid int,
	) (float32, error)
	GetAgentStates(
		ctx context.Context,
	) ([]models.Agent, error)
	GetExpressionsForUser(
		ctx context.Context,
		uid int,
	) ([]models.Expression, error)
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, error)
	Register(
		ctx context.Context,
		email string,
		password string,
	) (int, error)
}

func New(log *slog.Logger, storage httpservice.ExpressionStorage, port int, secret string, client *grpc.Client) *Server {
	httpService := httpservice.New(log, storage, client)

	server := &Server{
		log:         log,
		storage:     storage,
		secret:      secret,
		httpService: httpService,
	}
	serveMux := mux.NewRouter()

	// Хэндлеры для запросов с сайта
	serveMux.HandleFunc("/", server.MainPage)
	serveMux.Handle("/expression", middleware.ValidateToken(middleware.ValidateExpressionMiddleware(http.HandlerFunc(server.EvaluateExpression)), server.secret)).Methods("POST")
	serveMux.Handle("/expression", middleware.ValidateToken(http.HandlerFunc(server.GetExpressionById), server.secret)).Methods("GET")
	serveMux.Handle("/all_expressions", middleware.ValidateToken(http.HandlerFunc(server.GetExpressionsForUser), server.secret)).Methods("GET")
	serveMux.HandleFunc("/login", server.Login).Methods("POST")
	serveMux.HandleFunc("/register", server.Register).Methods("POST")
	serveMux.HandleFunc("/agents_state", server.GetAgentStates).Methods("GET")

	// Хэндлеры для API
	serveMux.HandleFunc("/api/v1/getimpodencekey", server.GetImpodenceKeyHandler).Methods("POST")

	http.Handle("/", serveMux)

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: serveMux,
	}

	server.server = s
	return server
}

func (s *Server) Run() error {
	const op = "server.Run"

	l, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = s.server.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
