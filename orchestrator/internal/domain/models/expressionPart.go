package models

import shuntingYard "github.com/a-romash/go-shunting-yard"

type ExpressionPart struct {
	FirstOperand  *shuntingYard.RPNToken      `json:"firstOperand"`
	SecondOperand *shuntingYard.RPNToken      `json:"secondOperand"`
	Operation     *shuntingYard.RPNToken      `json:"operation"`
	Duration      int                         `json:"duration"` // в секундах
	IdExpression  string                      `json:"id"`
	Result        chan *shuntingYard.RPNToken `json:"result"`
}

func NewExpressionPart(firstOperand, secondOperand, operation *shuntingYard.RPNToken, id string, duration int) *ExpressionPart {
	return &ExpressionPart{
		FirstOperand:  firstOperand,
		SecondOperand: secondOperand,
		Operation:     operation,
		Duration:      duration,
		IdExpression:  id,
		Result:        make(chan *shuntingYard.RPNToken),
	}
}
