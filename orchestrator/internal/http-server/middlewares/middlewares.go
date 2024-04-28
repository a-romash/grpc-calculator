package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	model "github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	expressionparser "github.com/a-romash/grpc-calculator/orchestrator/internal/lib/expressionParser"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/lib/jwt"
)

type myRequest struct {
	Expression string `json:"expression"`
	Id         string `json:"-"`
}

// Проверяем на валидность наше выражение и передаём уже паршенное (парсенное?) в Handler
func ValidateExpressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Парсим JSON
		var request myRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Error while parsing JSON", http.StatusBadRequest)
			return
		}

		// Парсим полученное выражение
		postfixExpression, err := expressionparser.ParseExpression(request.Expression)
		if err != nil {
			http.Error(w, "Error while parsing expression", http.StatusBadRequest)
			return
		}

		expression := model.Create(request.Expression, postfixExpression, expressionparser.CreateImpodenceKey(request.Expression))

		// Передаём в реквест контекст с выражением для дальнейшей работы с ним в хэндлере
		rWithContext := r.WithContext(context.WithValue(r.Context(), "expression", expression))
		// Пишем в хэдер статус код, что всё хорошо
		w.WriteHeader(http.StatusAccepted)
		next.ServeHTTP(w, rWithContext)
	})
}

// Проверка и валидация токена
func ValidateToken(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")
		token := strings.Split(bearer, " ")
		if len(token) != 2 {
			http.Error(w, "token is invalid! "+strings.Join(token, " "), http.StatusUnauthorized)
			return
		}
		if token[1] == "" {
			http.Error(w, "token is required!", http.StatusUnauthorized)
			return
		}
		id, err := jwt.ValidateToken(token[1], secret)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, jwt.ErrTokenExpired.Error(), http.StatusUnauthorized)
				return
			}
			http.Error(w, jwt.ErrUnexpectedSigningMethod.Error(), http.StatusInternalServerError)
			return
		}
		rWithCtx := r.WithContext(context.WithValue(r.Context(), "uid", id))

		next.ServeHTTP(w, rWithCtx)
	})
}

// Мидлварь для отлавливания паник
func PanicMiddleware(next *http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		(*next).ServeHTTP(w, r)
	})
}
