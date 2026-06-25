package items

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("item not found")

type ItemID uint64

func (id ItemID) String() string {
	return fmt.Sprintf("Nº%d", id)
}
