package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	shuntingYard "github.com/a-romash/go-shunting-yard"
	"github.com/a-romash/protos/gen/go/orchestrator"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	api orchestrator.OrchestratorClient
	log *slog.Logger
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string, // address SSO-server
	timeout time.Duration, // timeout of doing every try
	retriesCount int, // amount of retries
) (*Client, error) {
	const op = "grpc.New"

	log = log.With(
		slog.String("op", op),
	)

	log.Info("start creating client")

	// Опции для интерсептора grpcretry
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	// Опции для интерсептора grpclog
	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	// Создаём соединение с gRPC-сервером SSO для клиента
	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := orchestrator.NewOrchestratorClient(cc)

	log.Info("client created succesfully")

	return &Client{
		api: grpcClient,
		log: log,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Heartbeat(
	ctx context.Context,
	id int,
) error {
	const op = "grpc.Heartbeat"

	_, err := c.api.Heartbeat(ctx, &orchestrator.IsAlive{
		IsAlive: true,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GiveResultOfExpression(
	ctx context.Context,
	id_expression string,
	result float32,
	idAgent int,
) error {
	const op = "grpc.GiveResultOfExpression"

	_, err := c.api.GiveResultOfExpression(ctx, &orchestrator.ResultOfExpression{
		IdExpression: id_expression,
		Result:       result,
		IdAgent:      int32(idAgent),
	})
	if err != nil {
		c.log.Error(err.Error() + ". op: " + op)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func fromPrototokensToRPNTokens(prototokens []*orchestrator.RPNToken) ([]*shuntingYard.RPNToken, error) {
	var tokens []*shuntingYard.RPNToken

	for _, el := range prototokens {
		switch el.Type {
		case 1:
			float_value, err := strconv.ParseFloat(el.Value, 64)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, shuntingYard.NewRPNOperandToken(float_value))
		case 2:
			tokens = append(tokens, shuntingYard.NewRPNOperatorToken(el.Value))
		}
	}

	return tokens, nil
}

func (c *Client) GetExpressionToEvaluate(
	ctx context.Context,
	id_agent int,
) (string, []*shuntingYard.RPNToken, error) {
	const op = "grpc.GetExpressionToEvaluate"

	resp, err := c.api.GetExpressionToEvaluate(ctx, &orchestrator.IdAgent{
		IdAgent: int32(id_agent),
	})
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := fromPrototokensToRPNTokens(resp.PostfixExpression)
	if err != nil {
		return "", nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp.IdExpression, tokens, nil
}

func (c *Client) RegisterNewAgent(
	ctx context.Context,
) (int, error) {
	const op = "grpc.RegisterNewAgent"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("start registering")

	resp, err := c.api.RegisterNewAgent(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("registering was succesful")

	return int(resp.IdAgent), nil
}

func (c *Client) RemoveAgent(
	ctx context.Context,
	idAgent int,
) error {
	const op = "grpc.RemoveAgent"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("start removing")

	_, err := c.api.RemoveAgent(ctx, &orchestrator.IdAgent{
		IdAgent: int32(idAgent),
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info("removing was succesful!")
	return nil
}
