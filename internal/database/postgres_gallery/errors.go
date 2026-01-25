package postgres_gallery

import "errors"

var (
	ErrCouldNotInsert                 = errors.New(`could not insert character`)
	ErrCouldNotGet                    = errors.New(`could not get character`)
	ErrCouldNotFind                   = errors.New(`could not find character`)
	ErrFailedInitializeTransaction    = errors.New(`failed to initialize transaction`)
	ErrFailedCommitTransaction        = errors.New(`failed to commit transaction`)
	ErrFailedSelectCharacterInventory = errors.New(`failed to select character inventory`)
)
