package postgres

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	expressionparser "github.com/a-romash/grpc-calculator/orchestrator/internal/lib/expressionParser"
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
	CREATE TABLE IF NOT EXISTS expressions(
		id VARCHAR(255) PRIMARY KEY UNIQUE,
		expression VARCHAR(255) NOT NULL UNIQUE,
		result FLOAT,
		created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		solved_at TIMESTAMP,
		status VARCHAR(255) NOT NULL DEFAULT 'new'
	);

	CREATE TABLE IF NOT EXISTS agents(
		id SERIAL PRIMARY KEY,
		last_heartbeat TIMESTAMP NOT NULL,
		status VARCHAR(255) NOT NULL DEFAULT 'free'
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

// RegisterApp registers new app and returns id
func (db *Postgresql) RegisterApp(ctx context.Context, name, secret string) (int64, error) {
	const op = "storage.postgres.RegisterApp"

	const sql string = `
	INSERT INTO apps (name, secret)
	VALUES ($1, $2)
	`

	_, err := db.pool.Exec(ctx, sql, name, secret)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	const sql2 = `
	SELECT (id) FROM apps
	WHERE name = $1;
	`
	rows, _ := db.pool.Query(ctx, sql2, name)
	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.App])
	if errors.Is(err, pgx.ErrNoRows) {
		return -1, err
	} else if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	return int64(app.ID), nil
}

// Heartbeat updates last_heartbet of agent
func (db *Postgresql) Heartbeat(ctx context.Context, id_agent int) error {
	const op = "storage.postgres.Heartbeat"

	const sql = `
	UPDATE agents
	SET last_heartbeat = CURRENT_TIMESTAMP
	WHERE id=$1;
	`

	_, err := db.pool.Exec(ctx, sql, id_agent)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// RemoveAgent implements orch.ExpressionStorage.
func (db *Postgresql) RemoveAgent(ctx context.Context, id_agent int) error {
	const op = "storage.postgres.RemoveAgent"

	const sql = `
	DELETE FROM agents
	WHERE id = $1;
	`

	_, err := db.pool.Exec(ctx, sql, id_agent)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// SaveExpression saves expression to db and returns id of expression.
func (db *Postgresql) SaveExpression(ctx context.Context, expression string) (string, error) {
	const op = "storage.postgres.SaveExpression"

	const sql = `
	INSERT INTO expressions (id, expression)
	VALUES ($1, $2);
	`

	id := expressionparser.CreateImpodenceKey(expression)

	_, err := db.pool.Exec(ctx, sql, id, expression)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// SaveResult saves result of expression
func (db *Postgresql) SaveResult(ctx context.Context, idExpression string, result float32) error {
	const op = "storage.postgres.SaveResult"

	const sql = `
	UPDATE expressions
	SET result=$1
	WHERE id=$2;
	`
	_, err := db.pool.Exec(ctx, sql, result, idExpression)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetExpressionToEvaluate returns 1 'new' expression
func (db *Postgresql) GetExpressionToEvaluate(ctx context.Context, id_agent int) (*models.Expression, error) {
	const op = "storage.postgres.GetExpressionToEvaluate"

	const sql = `
	SELECT id, expression FROM expressions
	WHERE status = 'new';
	`

	row := db.pool.QueryRow(ctx, sql, id_agent)
	var id, expression string

	err := row.Scan(&id, &expression)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := expressionparser.ParseExpression(expression)
	if err != nil {
		const sql3 = `
		UPDATE expressions
		SET status='invalid'
		WHERE id=$2
		`
		_, err = db.pool.Exec(ctx, sql3, fmt.Sprintf("solving-%d", id_agent), id)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: %w", op, errors.New("expression is invalid"))
	}

	expressionModel := models.Create(expression, tokens, id)

	const sql2 = `
	UPDATE expressions
	SET status=$1
	WHERE id=$2
	`

	_, err = db.pool.Exec(ctx, sql2, fmt.Sprintf("solving-%d", id_agent), id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &expressionModel, nil
}

// RegisterNewAgent registers new agent and returns id
func (db *Postgresql) RegisterNewAgent(ctx context.Context) (int, error) {
	const op = "storage.postgres.RegisterNewAgent"

	const sql string = `
	INSERT INTO agents (last_heartbeat)
	VALUES (CURRENT_TIMESTAMP) 
	RETURNING id;
	`

	var id int

	row := db.pool.QueryRow(ctx, sql)
	err := row.Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// App returns app by id.
// func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
// 	const op = "storage.sqlite.App"

// 	stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
// 	if err != nil {
// 		return models.App{}, fmt.Errorf("%s: %w", op, err)
// 	}

// 	row := stmt.QueryRowContext(ctx, id)

// 	var app models.App
// 	err = row.Scan(&app.ID, &app.Name, &app.Secret)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
// 		}

// 		return models.App{}, fmt.Errorf("%s: %w", op, err)
// 	}

// 	return app, nil
// }
