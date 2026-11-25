package auth

import (
	"context"
	"dZev1/character-gallery/internal/metrics"
	"dZev1/character-gallery/internal/uuidv7"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PGErrUniqueViolation = "23505"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateUser      = errors.New("username or email already in use")
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

func (s *Service) Register(ctx context.Context, username, email, password string) (*UserResponse, error) {
	email = strings.ToLower(strings.TrimSpace(email))

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

	id, err := uuidv7.New()
	if err != nil {
		log.Println("Error generating user id:", err)
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	user, err := s.repo.SaveUser(ctx, tx, id, username, email, hash)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == PGErrUniqueViolation {
		return nil, ErrDuplicateUser
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
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: *user.CreatedAt,
	}, nil
}

func (s *Service) Login(ctx context.Context, identifier, password string) (*LoginResponse, error) {
	var (
		user *User
		err  error
	)
	if strings.Contains(identifier, "@") {
		user, err = s.repo.FindUserByEmail(ctx, s.pool, strings.ToLower(strings.TrimSpace(identifier)))
	} else {
		user, err = s.repo.FindUserByUsername(ctx, s.pool, identifier)
	}
	if err != nil {
		metrics.AuthLoginAttemptsTotal.WithLabelValues("invalid_credentials").Inc()
		return nil, ErrInvalidCredentials
	}

	if !CheckPassword(password, user.PasswordHash) {
		metrics.AuthLoginAttemptsTotal.WithLabelValues("invalid_credentials").Inc()
		return nil, ErrInvalidCredentials
	}

	token, expiresAt, err := generateJWT(user, s.jwtSecret)
	if err != nil {
		metrics.AuthLoginAttemptsTotal.WithLabelValues("error").Inc()
		log.Println("Error generating JWT:", err)
		return nil, fmt.Errorf("generate token: %w", err)
	}

	metrics.AuthLoginAttemptsTotal.WithLabelValues("success").Inc()
	return &LoginResponse{Token: token, ExpiresAt: expiresAt}, nil
}
