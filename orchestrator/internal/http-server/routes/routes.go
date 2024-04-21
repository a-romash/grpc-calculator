package routes

import (
	"net/http"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/handlers"
	middleware "github.com/a-romash/grpc-calculator/orchestrator/internal/http-server/middlewares"
	"github.com/gorilla/mux"
)

func New() *mux.Router {
	serveMux := mux.NewRouter()

	// Хэндлеры для запросов с сайта
	serveMux.HandleFunc("/", handlers.MainPage)
	serveMux.Handle("/expression", middleware.ValidateExpressionMiddleware(http.HandlerFunc(handlers.ExpressionHandler))).Methods("POST")
	serveMux.Handle("/agentstate", nil)
	serveMux.HandleFunc("/expression", handlers.GetExpressionById).Methods("GET")

	// Хэндлеры для API
	serveMux.HandleFunc("/api/v1/getimpodencekey", handlers.GetImpodenceKeyHandler).Methods("POST")

	http.Handle("/", serveMux)
	return serveMux
}
