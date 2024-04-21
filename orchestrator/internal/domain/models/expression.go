package models

import (
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

type Expression struct {
	InfinixExpression string                   `json:"infinix"`
	PostfixExpression []*shuntingYard.RPNToken `json:"postfix"`
	Result            float64                  `json:"result"`
	CreatedAt         time.Time                `json:"createdAt"`
	SolvedAt          time.Time                `json:"solvedAt"`
	Status            Status                   `json:"status"`
	IdExpression      string                   `json:"id"`
}

func Create(infinixExpression string, parsedExpression []*shuntingYard.RPNToken, id string) Expression {
	return Expression{
		InfinixExpression: infinixExpression,
		PostfixExpression: parsedExpression,
		CreatedAt:         time.Now(),
		IdExpression:      id,
		Status:            Solving,
	}
}

type Status string

const (
	Solved  Status = "completed"
	Solving Status = "solving"
	Invalid Status = "invalid"
)

const (
	Operation int = 2
	Operand   int = 1
)
