package postgres_gallery

import (
	"fmt"
	"log"
	"os"

	"dZev1/character-gallery/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresCharacterGallery(connStr string) (models.CharacterGallery, error) {
	var err error
	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection to database: %v", err)
	}

	schema, err := os.ReadFile("./internal/database/postgres_gallery/schema.sql")
	if err != nil {
		return nil, fmt.Errorf("could not load schema: %v", err)
	}

	db.MustExec(string(schema))

	log.Println("Database connection established")

	return &PostgresCharacterGallery{
		db: db,
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
