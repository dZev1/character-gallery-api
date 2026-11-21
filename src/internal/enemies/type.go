package enemies

type Type string

const (
	Normal   Type = "normal"
	Elite    Type = "elite"
	MiniBoss Type = "mini_boss"
	Boss     Type = "boss"
)

func (t Type) String() string {
	return string(t)
}

func (t Type) IsValid() bool {
	switch t {
	case Normal, Elite, MiniBoss, Boss:
		return true
	}
	return false
}
