package models

import (
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

type ExpressionPart struct {
	FirstOperand  *shuntingYard.RPNToken      `json:"firstOperand"`
	SecondOperand *shuntingYard.RPNToken      `json:"secondOperand"`
	Operation     *shuntingYard.RPNToken      `json:"operation"`
	Duration      time.Duration               `json:"duration"` // в секундах
	IdExpression  string                      `json:"id"`
	Result        chan *shuntingYard.RPNToken `json:"result"`
}

func NewExpressionPart(firstOperand, secondOperand, operation *shuntingYard.RPNToken, id string, duration time.Duration) *ExpressionPart {
	return &ExpressionPart{
		FirstOperand:  firstOperand,
		SecondOperand: secondOperand,
		Operation:     operation,
		Duration:      duration,
		IdExpression:  id,
		Result:        make(chan *shuntingYard.RPNToken),
	}
}
