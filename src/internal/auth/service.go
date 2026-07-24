package auth

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PGErrUniqueViolation = "23505"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrDuplicateUsername  = errors.New("username already taken")
)

type Service struct {
	repo      Repository
	pool      *pgxpool.Pool
	jwtSecret string
}

func NewService(repo Repository, pool *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		pool:      pool,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(ctx context.Context, username, password string) (*UserResponse, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	hash, err := HashPassword(password)
	if err != nil {
		log.Println("Error creating user:", err)
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.SaveUser(ctx, tx, username, hash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
		return nil, ErrDuplicateUsername
	}
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("Error committing transaction:", err)
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: *user.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	user, err := s.repo.FindUserByUsername(ctx, s.pool, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !CheckPassword(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := generateJWT(user, s.jwtSecret)
	if err != nil {
		log.Println("Error generating JWT:", err)
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{Token: token, ExpiresAt: expiresAt}, nil
}
