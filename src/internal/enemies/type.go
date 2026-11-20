package enemies

type Type string

const (
	Normal   Type = "Normal"
	Elite         = "Elite"
	MiniBoss      = "MiniBoss"
	Boss          = "Boss"
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
