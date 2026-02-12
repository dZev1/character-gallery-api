package postgres_gallery

import (
	"dZev1/character-gallery/models/auth"
	"github.com/jmoiron/sqlx"
)

type PGAuthStore struct {
	db *sqlx.DB
}

func NewAuthStore(db *sqlx.DB) auth.AuthStore {
	return &PGAuthStore{
		db: db,
	}
}

func (s *PGAuthStore) ValidateAPIKey(keyHash string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(SELECT 1 FROM api_keys WHERE key_hash = $1)
	`

	err := s.db.QueryRowx(query, keyHash).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *PGAuthStore) UpdateLastUsed(keyHash string) error {
	query := `
		UPDATE api_keys SET last_used = NOW() WHERE key_hash = $1
	`

	_, err := s.db.Exec(query, keyHash)
	if err != nil {
		return err
	}

	return nil
}

func (s *PGAuthStore) CreateAPIKey(name string) (string, error) {
	keyHash, rawKey, err := auth.GenerateAPIKey()
	if err != nil {
		return "", err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO api_keys (name, key_hash)
		VALUES ($1, $2)
	`

	_, err = tx.Exec(query, name, keyHash)
	if err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return rawKey, nil
}