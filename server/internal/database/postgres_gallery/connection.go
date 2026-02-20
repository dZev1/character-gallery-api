package postgres_gallery

import (
	_ "embed"
	"fmt"
	"log"

	"dZev1/character-gallery/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

//go:embed schema.sql
var schemaSQL string

func NewPostgresCharacterGallery(connStr string) (models.CharacterGallery, error) {
	var err error
	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection to database: %v", err)
	}

	db.MustExec(string(schemaSQL))

	log.Println("Database connection established")

	return &PostgresCharacterGallery{
		db:        db,
		AuthStore: NewAuthStore(db),
	}, nil
}

func (cg *PostgresCharacterGallery) Close() error {
	log.Println("Database connection terminated")
	err := cg.db.Close()
	if err != nil {
		return fmt.Errorf("error closing database connection: %v\n", err)
	}
	return nil
}
