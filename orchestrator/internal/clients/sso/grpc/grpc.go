package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-romash/protos/gen/go/sso"
	"google.golang.org/grpc"
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

	cc, err := grpc.DialContext(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := sso.NewAuthClient(cc)

	return &Client{
		api:    grpcClient,
		log:    log,
		app_id: app_id,
	}, nil
}

func (c *Client) Register(
	ctx context.Context,
	email string,
	password string,
) (int, error) {
	const op = "grpc.Register"

	resp, err := c.api.Register(ctx, &sso.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

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
