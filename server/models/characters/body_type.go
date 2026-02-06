package characters

type BodyType string

const (
	TypeA BodyType = "type_a"
	TypeB BodyType = "type_b"
)

func (bt BodyType) String() string {
	return string(bt)
}

func (bt BodyType) Validate() bool {
	switch bt {
	case TypeA, TypeB:
		return true
	}
	return false
} 