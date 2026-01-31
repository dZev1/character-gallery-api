package characters

import "fmt"

type CharacterID uint64

func (id CharacterID) String() string {
	return fmt.Sprintf("Nº%d", id)
}
