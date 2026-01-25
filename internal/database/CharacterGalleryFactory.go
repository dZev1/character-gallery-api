package database

import (
	"fmt"
	"github.com/dZev1/character-gallery/models"
	"github.com/dZev1/character-gallery/internal/database/postgres_gallery"
)

func NewCharacterGallery(dbType string, connectionString string) (models.CharacterGallery, error) {
	switch dbType {
	case "postgres":
		return postgres_gallery.NewPostgresCharacterGallery(connectionString)
	// case "mariadb":
	// 	return mariadb_gallery.NewMariaDBCharacterGallery(connectionString)
	// case "sqlite":
	// 	return sqlite_gallery.NewSQLiteCharacterGallery(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}