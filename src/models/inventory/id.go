package inventory

import "fmt"

type ItemID uint64

func (id ItemID) String() string {
	return fmt.Sprintf("Nº%d", id)
}
