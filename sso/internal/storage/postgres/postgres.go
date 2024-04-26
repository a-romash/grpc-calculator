package postgres

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/a-romash/grpc-calculator/sso/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type Postgresql struct {
	pool      *pgxpool.Pool
	closeOnce sync.Once
}

func Connect(databaseUrl string) (db *Postgresql, err error) {
	config := Config(databaseUrl)
	return ConnectWithConfig(config)
}

func ConnectWithConfig(config *pgxpool.Config) (db *Postgresql, err error) {
	for i := 0; i < 5; i++ {
		p, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil || p == nil {
			time.Sleep(3 * time.Second)
			continue
		}
		log.Printf("pool returned from connect: idk from where so i am really lazy for normal logs tho")
		db = &Postgresql{
			pool: p,
		}
		err = Init(db.pool)
		if err != nil {
			slog.Error("error initing database")
			return nil, err
		}
		slog.Info("database was successfully init")
		return db, nil
	}
	err = errors.Wrap(err, "timed out waiting to connect postgres")
	slog.Error("timed out waiting to connect postgres")
	return nil, err
}

func (db *Postgresql) Close() {
	db.closeOnce.Do(func() {
		db.pool.Close()
	})
}

func Config(databaseUrl string) *pgxpool.Config {
	const defaultMaxConns = int32(10)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	dbConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		log.Fatal("Failed to create a config, error: ", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		slog.Debug("Before acquiring the connection pool to the database!!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		slog.Debug("After releasing the connection pool to the database!!")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		slog.Debug("Closed the connection pool to the database!!")
	}

	return dbConfig
}

func Init(p *pgxpool.Pool) (err error) {
	const sql string = `
	CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		pass_hash BYTEA NOT NULL
	);

	CREATE TABLE IF NOT EXISTS apps(
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		secret VARCHAR(255) NOT NULL 
	);
	`

	_, err = p.Exec(context.Background(), sql)
	return err
}

func (db *Postgresql) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	const sql = `
	INSERT INTO users (email, pass_hash)
  	VALUES ($1, $2);
	`

	_, err := db.pool.Exec(ctx, sql, email, passHash)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	const sql2 = `
	SELECT (id) FROM users
	WHERE email = $1;
	`
	rows, _ := db.pool.Query(ctx, sql2, email)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return -1, err
	} else if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return int64(user.ID), nil
}

func (db *Postgresql) User(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.postgres.User"

	const sql string = `
	SELECT (id, email, pass_hash) FROM users
	WHERE email = $1;
	`

	rows, _ := db.pool.Query(ctx, sql, email)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

func (db *Postgresql) App(ctx context.Context, id int) (*models.App, error) {
	const op = "storage.postgres.App"

	const sql string = `
	SELECT (id, name, secret) FROM apps
	WHERE id = $1;
	`

	rows, _ := db.pool.Query(ctx, sql, id)
	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.App])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &app, nil
}
