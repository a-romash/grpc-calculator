package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-romash/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api    sso.AuthClient
	log    *slog.Logger
	app_id int
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string, // address SSO-server
	timeout time.Duration, // timeout of doing every try
	retriesCount int, // amount of retries
	app_id int, // id of application
) (*Client, error) {
	const op = "grpc.New"

	log = log.With(
		slog.String("op", op),
		slog.Int64("app_id", int64(app_id)),
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

	grpcClient := sso.NewAuthClient(cc)

	log.Info("client created succesfully")

	return &Client{
		api:    grpcClient,
		log:    log,
		app_id: app_id,
	}, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) Register(
	ctx context.Context,
	email string,
	password string,
) (int, error) {
	const op = "grpc.Register"

	log := c.log.With(
		slog.String("op", op),
	)

	log.Info("start registering")

	resp, err := c.api.Register(ctx, &sso.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("registering was succesful")

	return int(resp.UserId), nil
}

func (c *Client) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "grpc.Login"

	resp, err := c.api.Login(ctx, &sso.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    int32(c.app_id),
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.Token, nil
}
