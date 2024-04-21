package orch

import (
	"context"
	"log/slog"
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
)

type Orchestrator struct {
	log     *slog.Logger
	storage ExpressionStorage
}

type ExpressionStorage interface {
	SaveExpression(
		ctx context.Context,
		expression string,
		result float32,
		created time.Time,
		solved time.Time,
	) error
}

func New(
	log *slog.Logger,
	storage ExpressionStorage,
) *Orchestrator {
	return &Orchestrator{
		log:     log,
		storage: storage,
	}
}

// TODO: сделать
func (o *Orchestrator) Heartbeat(
	ctx context.Context,
	is_alive bool,
	id_agent int,
) error {
	return nil
}

func (o *Orchestrator) GetExpressionToEvaluate(
	ctx context.Context,
	id_agent int,
) (string, []shuntingYard.RPNToken, error) {
	return "", []shuntingYard.RPNToken{}, nil
}

func (o *Orchestrator) SaveResultOfExpression(
	ctx context.Context,
	id_expression string,
	result float32,
) error {
	return nil
}
