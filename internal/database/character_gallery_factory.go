package database

import (
	"fmt"
	"github.com/dZev1/character-gallery/internal/database/postgres_gallery"
	"github.com/dZev1/character-gallery/models"
)

func NewCharacterGallery(dbType string, connectionString string) (models.CharacterGallery, error) {
	switch dbType {
	case "postgres":
		return postgres_gallery.NewPostgresCharacterGallery(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
