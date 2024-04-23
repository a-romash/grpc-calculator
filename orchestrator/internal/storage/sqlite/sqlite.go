package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/a-romash/grpc-calculator/orchestrator/internal/domain/models"
	expressionparser "github.com/a-romash/grpc-calculator/orchestrator/internal/lib/expressionParser"
	"github.com/a-romash/grpc-calculator/orchestrator/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	const sql = `
	CREATE TABLE IF NOT EXISTS expressions(
		id TEXT PRIMARY KEY,
		expression TEXT NOT NULL UNIQUE,
		result FLOAT,
		created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		solved_at DATETIME,
		status TEXT NOT NULL DEFAULT "new"
	);
	CREATE INDEX IF NOT EXISTS idx_expressions_ids ON expressions(id);
	  
	CREATE TABLE IF NOT EXISTS agents(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		last_heartbeat DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT "free"
	);

	CREATE TABLE IF NOT EXISTS apps(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		secret TEXT NOT NULL
	);
	`

	stmt, err := db.Prepare(sql)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	return s.db.Close()
}

// RegisterApp registers new app and returns id
func (s *Storage) RegisterApp(ctx context.Context, name, secret string) (int64, error) {
	const op = "storage.sqlite.RegisterApp"

	stmt, err := s.db.Prepare("INSERT INTO apps(name, secret) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, name, secret)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAppExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// SaveExpression saves expression to db and returns id of expression.
func (s *Storage) SaveExpression(ctx context.Context, expression string) (string, error) {
	const op = "storage.sqlite.SaveExpression"

	stmt, err := s.db.Prepare("INSERT INTO expressions(id, expression) VALUES(?, ?)")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	id := expressionparser.CreateImpodenceKey(expression)

	_, err = stmt.ExecContext(ctx, id, expression)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

// Heartbeat updates last_heartbet of agent
func (s *Storage) Heartbeat(ctx context.Context, id_agent int) error {
	const op = "storage.sqlite.Heartbeat"

	stmt, err := s.db.Prepare("UPDATE agents SET last_heartbeat = CURRENT_TIMESTAMP WHERE id = (?);")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, id_agent)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetExpressionToEvaluate returns 1 'new' expression
func (s *Storage) GetExpressionToEvaluate(ctx context.Context, id_agent int) (*models.Expression, error) {
	const op = "storage.sqlite.GetExpressionToEvaluate"

	stmt, err := s.db.Prepare("SELECT id, expression FROM expressions WHERE status = 'new'")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	res := stmt.QueryRowContext(ctx)

	var id, expression string
	err = res.Scan(&id, &expression)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tokens, err := expressionparser.ParseExpression(expression)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, errors.New("expression is invalid"))
	}

	expressionModel := models.Create(expression, tokens, id)
	return &expressionModel, nil
}

// SaveResult saves result of expression to db
func (s *Storage) SaveResult(ctx context.Context, id_expression string, result float32) error {
	const op = "storage.sqlite.SaveResult"

	stmt, err := s.db.Prepare("UPDATE expressions SET result = (?), solved_at = CURRENT_TIMESTAMP, status = 'solved' WHERE id = (?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, result, id_expression)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// RemoveAgent implements orch.ExpressionStorage.
func (s *Storage) RemoveAgent(ctx context.Context, id_agent int) error {
	return errors.New("unimplemented")
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
