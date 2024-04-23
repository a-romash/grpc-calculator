package orch

import (
	"context"
	"fmt"
	"log/slog"

	shuntingYard "github.com/a-romash/go-shunting-yard"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
)

type Orchestrator struct {
	log     *slog.Logger
	storage ExpressionStorage
}

type ExpressionStorage interface {
	SaveExpression(
		ctx context.Context,
		expression string,
	) (string, error)
	Heartbeat(
		ctx context.Context,
		id_agent int,
	) error
	RemoveAgent(
		ctx context.Context,
		id_agent int,
	) error
	GetExpressionToEvaluate(
		ctx context.Context,
		id_agent int,
	) (*models.Expression, error)
	SaveResult(
		ctx context.Context,
		id_expression string,
		result float32,
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
	const op = "Orch.Heartbeat"

	if is_alive {
		if err := o.storage.Heartbeat(ctx, id_agent); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		return nil
	}

	// executes when agent isn't alive
	if err := o.storage.RemoveAgent(ctx, id_agent); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *Orchestrator) GetExpressionToEvaluate(
	ctx context.Context,
	id_agent int,
) (string, []*shuntingYard.RPNToken, error) {
	const op = "Orch.GetExpressionToEvaluate"

	expression, err := o.storage.GetExpressionToEvaluate(ctx, id_agent)
	if err != nil {
		return "", []*shuntingYard.RPNToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return expression.IdExpression, expression.PostfixExpression, nil
}

func (o *Orchestrator) SaveResultOfExpression(
	ctx context.Context,
	id_expression string,
	result float32,
) error {
	const op = "Orch.SaveResult"

	if err := o.storage.SaveResult(ctx, id_expression, result); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
