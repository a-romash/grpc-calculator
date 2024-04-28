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
		expression VARCHAR(255) NOT NULL,
		uid SERIAL NOT NULL REFERENCES users(id),
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
		name VARCHAR(255) NOT NULL UNIQUE,
		secret VARCHAR(255) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		pass_hash BYTEA NOT NULL
	);
	`

	_, err = p.Exec(context.Background(), sql)
	return err
}

// RegisterApp registers new app and returns id
func (db *Postgresql) RegisterApp(ctx context.Context, name, secret string) (int64, error) {
	const op = "storage.postgres.RegisterApp"

	const sql string = `
	DELETE FROM apps
	WHERE name = $1;
	`

	_, err := db.pool.Exec(ctx, sql, name)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}

	const sql2 string = `
	INSERT INTO apps (name, secret)
	VALUES ($1, $2)
	RETURNING id; 
	`

	var id int64

	row := db.pool.QueryRow(ctx, sql2, name, secret)
	err = row.Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
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

	const sql2 = `
	UPDATE expressions
	SET status = 'new'
	WHERE status = $1;
	`

	_, err = db.pool.Exec(ctx, sql2, fmt.Sprintf("solving-%d", id_agent))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// SaveExpression saves expression to db and returns id of expression.
func (db *Postgresql) SaveExpression(ctx context.Context, expression *models.Expression, uid int) (string, error) {
	const op = "storage.postgres.SaveExpression"

	const sql = `
	INSERT INTO expressions (id, expression, uid)
	VALUES ($1, $2, $3);
	`

	_, err := db.pool.Exec(ctx, sql, expression.IdExpression, expression.InfinixExpression, uid)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return expression.IdExpression, nil
}

// SaveResult saves result of expression
func (db *Postgresql) SaveResult(ctx context.Context, idExpression string, result float32, idAgent int) error {
	const op = "storage.postgres.SaveResult"

	const sql = `
	UPDATE expressions
	SET result=$1, status='solved', solved_at=CURRENT_TIMESTAMP
	WHERE id=$2;
	`
	_, err := db.pool.Exec(ctx, sql, result, idExpression)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	const sql2 = `
	UPDATE agents
	SET status = 'free'
	WHERE id = $1;
	`

	_, err = db.pool.Exec(ctx, sql2, idAgent)
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

	row := db.pool.QueryRow(ctx, sql)
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
		WHERE id=$2;
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
	WHERE id=$2;
	`

	_, err = db.pool.Exec(ctx, sql2, fmt.Sprintf("solving-%d", id_agent), id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	const sql3 = `
	UPDATE agents
	SET status=$1
	WHERE id=$2;
	`

	_, err = db.pool.Exec(ctx, sql3, fmt.Sprintf("solving-%s", id), id_agent)
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

// GetAllAgent returns all agents from database
func (db *Postgresql) GetAllAgent(ctx context.Context) ([]models.Agent, error) {
	const op = "storage.postgres.GetAllAgent"

	const sql string = `
	SELECT * FROM agents;
	`

	rows, err := db.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	agents, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Agent])
	if err != nil {
		return nil, err
	}
	return agents, nil
}

// GetExpressionsForUser returns all expressions (solved and not) for exactly user
func (db *Postgresql) GetExpressionsForUser(ctx context.Context, uid int) ([]models.Expression, error) {
	const op = "storage.postgres.GetExpressionsForUser"

	const sql = `
	SELECT * FROM expressions
	WHERE uid = $1;
	`

	rows, err := db.pool.Query(ctx, sql, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	expressions, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[models.Expression])
	if err != nil {
		return nil, err
	}
	return expressions, nil
}

func (db *Postgresql) GetExpressionById(ctx context.Context, id string, uid int) (*models.Expression, error) {
	const op = "storage.postgres.GetExpressionById"

	const sql = `
	SELECT * FROM expressions
	WHERE id = $1 AND uid = $2;
	`

	rows, err := db.pool.Query(ctx, sql, id, uid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	expressions, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Expression])
	if err != nil {
		return nil, err
	}
	if len(expressions) == 0 {
		return nil, pgx.ErrNoRows
	}
	expression := &expressions[0]
	return expression, nil
}

// GetResultOfExpression returns result of exactly expression
func (db *Postgresql) GetResultOfExpression(ctx context.Context, id string) (float32, error) {
	const op = "storage.postgres.GetResultOfExpression"

	const sql = `
	SELECT (result) FROM expressions
	WHERE id = $1 AND status = 'solved';
	`

	row := db.pool.QueryRow(ctx, sql, id)

	var result float32
	err := row.Scan(&result)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}
