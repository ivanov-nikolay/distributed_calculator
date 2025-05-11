package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ivanov-nikolay/distributed_calculator/internal/app/entity"
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/repository"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	absPath := filepath.Join(".", path)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %v", err)
	}

	conn, err := sql.Open("sqlite3", absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v\n", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v\n", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := migrate(conn); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v\n", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			login TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS expressions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			expression TEXT NOT NULL,
			result FLOAT,
			status TEXT DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)
	return err
}

type ExpressionRepository struct {
	db *DB
}

func NewExpressionRepository(db *DB) repository.ExpressionRepository {
	return &ExpressionRepository{db: db}
}

func (r *ExpressionRepository) Add(ctx context.Context, userID int, expr string) (int, error) {
	res, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO expressions (user_id, expression) VALUES (?, ?)",
		userID, expr,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert expression: %v\n", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %v\n", err)
	}

	return int(id), nil
}

func (r *ExpressionRepository) UpdateResult(ctx context.Context, id int, result string) error {
	_, err := r.db.conn.ExecContext(ctx,
		"UPDATE expressions SET result = ?, status = 'completed', updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		result, id,
	)
	return err
}

func (r *ExpressionRepository) GetAll(ctx context.Context, userID int) ([]entity.Expression, error) {
	rows, err := r.db.conn.QueryContext(ctx,
		"SELECT id, expression as expr, result, status, user_id FROM expressions WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query expressions: %v\n", err)
	}
	defer rows.Close()

	var expressions []entity.Expression
	for rows.Next() {
		var expr entity.Expression
		var result sql.NullFloat64
		if err := rows.Scan(&expr.ID, &expr.Expr, &result, &expr.Status, &expr.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan expression: %v\n", err)
		}
		if result.Valid {
			expr.Result = &result.Float64
		}
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

func (r *ExpressionRepository) GetByID(ctx context.Context, userID int, id int) (entity.Expression, error) {
	var expr entity.Expression
	var result sql.NullFloat64

	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, expression as expr, result, status, user_id FROM expressions WHERE user_id = ? AND id = ?",
		userID, id,
	).Scan(&expr.ID, &expr.Expr, &result, &expr.Status, &expr.UserID)

	if err != nil {
		return entity.Expression{}, fmt.Errorf("failed to get expression: %v\n", err)
	}

	if result.Valid {
		expr.Result = &result.Float64
	}

	return expr, nil
}

func (r *ExpressionRepository) UpdateStatus(ctx context.Context, exprID string, status string) error {
	_, err := r.db.conn.ExecContext(ctx,
		"UPDATE expressions SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, exprID,
	)
	return err
}

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) repository.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, login, password string) error {
	_, err := r.db.conn.ExecContext(ctx,
		"INSERT INTO users (login, password) VALUES (?, ?)",
		login, password,
	)
	return err
}

func (r *UserRepository) GetID(ctx context.Context, login string) (int, error) {
	var id int
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id FROM users WHERE login = ?",
		login,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) Authenticate(ctx context.Context, login, password string) (bool, error) {
	var storedPassword string
	err := r.db.conn.QueryRowContext(ctx,
		"SELECT password FROM users WHERE login = ?",
		login,
	).Scan(&storedPassword)

	if err != nil {
		return false, err
	}

	return password == storedPassword, nil
}

func (r *UserRepository) Login(ctx context.Context, login, password string) (int, error) {
	var (
		userID         int
		storedPassword string
	)

	err := r.db.conn.QueryRowContext(ctx,
		"SELECT id, password FROM users WHERE login = ?;",
		login,
	).Scan(&userID, &storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("database error: %v", err)
	}

	//if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
	//	return 0, fmt.Errorf("invalid password")
	//}

	return userID, nil
}
