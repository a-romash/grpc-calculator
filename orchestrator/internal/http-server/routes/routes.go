package routes

import (
	"log/slog"
	"net/http"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/handlers"
	middleware "github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/middlewares"
	"github.com/gorilla/mux"
)

func New(log *slog.Logger, storage handlers.ExpressionStorage) *mux.Router {
	serveMux := mux.NewRouter()

	h := handlers.New(log, storage)

	// Хэндлеры для запросов с сайта
	serveMux.HandleFunc("/", h.MainPage)
	serveMux.Handle("/expression", middleware.ValidateExpressionMiddleware(http.HandlerFunc(h.ExpressionHandler))).Methods("POST")
	serveMux.Handle("/agentstate", nil)
	serveMux.HandleFunc("/expression", h.GetExpressionById).Methods("GET")

	// Хэндлеры для API
	serveMux.HandleFunc("/api/v1/getimpodencekey", h.GetImpodenceKeyHandler).Methods("POST")

	http.Handle("/", serveMux)
	return serveMux
}
