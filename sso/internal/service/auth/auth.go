package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/a-romash/grpc-calculator/sso/internal/domain/models"
	"github.com/a-romash/grpc-calculator/sso/internal/lib/jwt"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log      *slog.Logger
	storage  UserStorage
	tokenTTL time.Duration
}

type UserStorage interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
	User(ctx context.Context, email string) (models.User, error)
	App(ctx context.Context, appId int) (models.App, error)
}

func New(
	log *slog.Logger,
	storage UserStorage,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:      log,
		storage:  storage,
		tokenTTL: tokenTTL,
	}
}

// Login checks if user exists in database with given credentials and returns JWT
//
// If user exists but password nor email doesnt exist, returns error
// If user doesnt exists, return error
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"

	user, err := a.storage.User(ctx, email)
	if err != nil {
		// TODO: handle non-existent user
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		// executes when passwords doesnt matches
		return "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := a.storage.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// Register checks if user with given email exists and returns userID
// If user with given email already exists, returns error
func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "Auth.RegisterNewUser"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.storage.SaveUser(ctx, email, passHash)
	if err != nil {
		// TODO: handle if user with given email already exists\
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
