package models

import (
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

type Expression struct {
	InfinixExpression string                   `json:"infinix" db:"expression"`
	PostfixExpression []*shuntingYard.RPNToken `json:"postfix" db:"-"`
	Result            float64                  `json:"result" db:"result"`
	UserId            int                      `json:"uid" db:"uid"`
	CreatedAt         time.Time                `json:"createdAt" db:"created"`
	SolvedAt          time.Time                `json:"solvedAt" db:"solved_at"`
	Status            Status                   `json:"status" db:"status"`
	IdExpression      string                   `json:"id" db:"id"`
}

func Create(infinixExpression string, parsedExpression []*shuntingYard.RPNToken, id string) Expression {
	return Expression{
		InfinixExpression: infinixExpression,
		PostfixExpression: parsedExpression,
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
