package httpservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/clients/sso/grpc"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	"github.com/jackc/pgx/v5"
)

type HttpService struct {
	log     *slog.Logger
	storage ExpressionStorage
	client  *grpc.Client
}

type ExpressionStorage interface {
	RegisterApp(
		ctx context.Context,
		name string,
		secret string,
	) (int64, error)
	GetAllAgent(
		ctx context.Context,
	) ([]models.Agent, error)
	GetExpressionsForUser(
		ctx context.Context,
		uid int,
	) ([]models.Expression, error)
	GetExpressionById(
		ctx context.Context,
		id string,
		uid int,
	) (*models.Expression, error)
	SaveExpression(
		ctx context.Context,
		expression *models.Expression,
		uid int,
	) (string, error)
	GetResultOfExpression(
		ctx context.Context,
		id string,
	) (float32, error)
}

func New(
	log *slog.Logger,
	storage ExpressionStorage,
	client *grpc.Client,
) *HttpService {
	return &HttpService{
		log:     log,
		storage: storage,
		client:  client,
	}
}

func (s *HttpService) GetExpressionById(
	ctx context.Context,
	id string,
	uid int,
) (*models.Expression, error) {
	const op = "httpservice.GetExpressionById"

	log := s.log.With(
		slog.String("op", op),
		slog.String("id_expression", id),
		slog.Int("uid", uid),
	)

	log.Info("start getting expression")

	expression, err := s.storage.GetExpressionById(ctx, id, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		log.Error(err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("getting expression was succesful")

	return expression, err
}

func (s *HttpService) EvaluateExpression(
	ctx context.Context,
	expression *models.Expression,
	uid int,
) (float32, error) {
	const op = "httpservice.EvaluateExpression"

	log := s.log.With(
		slog.String("op", op),
		slog.String("id_expression", expression.IdExpression),
		slog.Int("uid", uid),
	)

	log.Info("start evaluating expression")
	id, err := s.storage.SaveExpression(ctx, expression, uid)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}
	var result float32
	for {
		result, err = s.storage.GetResultOfExpression(ctx, id)
		if err == nil {
			break
		}

		time.Sleep(10 * time.Second)
	}
	log.Info("expression evaluated!")
	return result, nil
}

func (s *HttpService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "httpservice.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("start logining")

	token, err := s.client.Login(context.Background(), email, password)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	log.Info("logining was succesful")
	return token, nil
}

func (s *HttpService) Register(
	ctx context.Context,
	email string,
	password string,
) (int, error) {
	const op = "httpservice.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("start registering")

	id, err := s.client.Register(context.Background(), email, password)
	if err != nil {
		log.Error(err.Error())
		return -1, err
	}

	log.Info("registering was succesful")
	return id, nil
}

func (s *HttpService) GetExpressionsForUser(
	ctx context.Context,
	uid int,
) ([]models.Expression, error) {
	const op = "httpservice.GetExpressionsForUser"

	log := s.log.With(
		slog.String("op", op),
		slog.Int("uid", uid),
	)

	log.Info("start getting expressions")

	expressions, err := s.storage.GetExpressionsForUser(ctx, uid)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("getting expressions was succesful!")
	return expressions, nil
}

func (s *HttpService) GetAgentStates(
	ctx context.Context,
) ([]models.Agent, error) {
	const op = "httpservice.GetAgentStates"

	log := s.log.With(
		slog.String("op", op),
	)

	log.Info("start getting agents")

	agent, err := s.storage.GetAllAgent(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("getting agents was succesful!")
	return agent, nil
}
