package match

const (
	FloorCeilType = iota
)

type Match interface {
	Type() int
}

// FloorCeil match that has a range of activation (start stop limits)
// while both params are optional, only one at a time can be ommited
type FloorCeil struct {
	Floor uint
	Ceil  uint
}

func (fc FloorCeil) Type() int {
	return FloorCeilType
}

func (fc FloorCeil) Overlaps(a FloorCeil) bool {
	ma := max([]uint{a.Floor, fc.Floor})
	mi := min([]uint{a.Ceil, fc.Ceil})
	return ma <= mi
}

func min(v []uint) (m uint) {
	if len(v) > 0 {
		m = v[0]
	}
	for i := 1; i < len(v); i++ {
		if v[i] < m {
			m = v[i]
		}
	}
	return
}

func max(v []uint) (m uint) {
	if len(v) > 0 {
		m = v[0]
	}
	for i := 1; i < len(v); i++ {
		if v[i] > m {
			m = v[i]
		}
	}
	return
}
