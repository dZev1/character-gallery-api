package characters

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("character not found")

type CharacterID uint64

func (id CharacterID) String() string {
	return fmt.Sprintf("Nº%d", id)
}
