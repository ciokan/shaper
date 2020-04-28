package markers

// Defines a marker with its value and parameters
type Marker struct {
	Identifier string
	Value      int
}

// Store for managing and allocating packet markers
// The idea is to recycle markers if the parameters are the same
type Markers struct {
	markers []Marker
}

var mObj *Markers

func New() *Markers {
	if mObj == nil {
		mObj = &Markers{}
	}
	return mObj
}

func (mm *Markers) Get(identifier string) (found bool, marker *Marker) {
	// create a new one if it doesn't exist (and add it to store)
	// otherwise return the existing one
	for _, m := range mm.markers {
		if m.Identifier == identifier {
			return true, &m
		}
	}
	m := Marker{
		Identifier: identifier,
		Value:      mm.genVal(),
	}
	mm.markers = append(mm.markers, m)
	return false, &m
}

func (mm *Markers) genVal() int {
	var existing []int
	for _, m := range mm.markers {
		existing = append(existing, m.Value)
	}
	return max(existing) + 1
}

func max(v []int) (m int) {
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
