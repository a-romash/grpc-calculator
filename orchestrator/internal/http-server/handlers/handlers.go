package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	model "github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	expressionparser "github.com/a-romash/grpc-calculator/orchestrator/internal/lib/expressionParser"
)

type myRequest struct {
	Expression string `json:"expression"`
	Id         string `json:"-"`
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, use curl :)")
}

func GetExpressionById(w http.ResponseWriter, r *http.Request) {
	_ = r.URL.Query().Get("id")

	// logic getting expression from db

	fmt.Fprintln(w, "Expression doesn't exist")
}

func ExpressionHandler(w http.ResponseWriter, r *http.Request) {
	expression := r.Context().Value("expression").(model.Expression)

	// // Проверяем выражение на наличие результата в базе данных (в ином случае отправляем агенту на вычисление)
	// // if el, ok := cache.Get(expression.IdExpression); ok {
	// // 	fmt.Fprintln(w, el.Result)
	// // 	return
	// // }

	// expression, err := orchestrator.SolveExpression(&expression)
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// // записываем в бд
	fmt.Fprintln(w, expression.Result)
}

// вообще по идее оно должно создаваться на фронтэнде, но т.к пока нет фронта - создаём на бэкенде (по запросу с фронта)
func GetImpodenceKeyHandler(w http.ResponseWriter, r *http.Request) {
	// Парсим JSON
	var request myRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Error while parsing JSON", http.StatusBadRequest)
		return
	}

	request.Expression = strings.ReplaceAll(request.Expression, " ", "")

	// и получаем ключ
	key := expressionparser.CreateImpodenceKey(request.Expression)

	w.Write([]byte(key))
}
