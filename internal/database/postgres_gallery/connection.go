package database

import (
	"fmt"
	"log"
	"os"

	"github.com/dZev1/character-gallery/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func NewPostgresCharacterGallery(connStr string, schemaPath string) (models.CharacterGallery, error) {
	var err error
	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not establish connection to database: %v", err)
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("could not load schema: %v", err)
	}

	db.MustExec(string(schema))

	log.Println("Database connection established")

	return &PostgresCharacterGallery{
		db: db,
	}, nil
}

func (cg *PostgresCharacterGallery) Close() {
	log.Println("Database connection terminated")
	err := cg.db.Close()
	if err != nil {
		log.Printf("error closing database connection: %v\n", err)
	}
}
