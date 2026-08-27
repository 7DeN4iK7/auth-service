package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/7DeN4iK7/auth-service/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUniqueViolation = errors.New("unique key violation")
	ErrUserNotFound    = errors.New("user not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(cfg config.PostgresConfig) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) CreateUser(ctx context.Context, req CreateUserParams) (int, error) {
	var id int

	err := r.pool.QueryRow(
		ctx,
		`
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2) 
		RETURNING id
		`,
		req.Username, req.PasswordHash,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrUniqueViolation
		}

		return 0, err
	}

	return id, err
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var user User

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT id, password_hash, username, created_at
		FROM users
		WHERE username = $1		
		`,
		username,
	).Scan(&user.ID, &user.PasswordHash, &user.Username, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int) (User, error) {
	var user User

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT id, password_hash, username, created_at
		FROM users
		WHERE id = $1		
		`,
		id,
	).Scan(&user.ID, &user.PasswordHash, &user.Username, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return user, nil
}
