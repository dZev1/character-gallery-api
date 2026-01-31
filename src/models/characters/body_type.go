package characters

type BodyType string

const (
	TYPE_A BodyType = "type_a"
	TYPE_B BodyType = "type_b"
)

func (bt BodyType) String() string {
	return string(bt)
}
