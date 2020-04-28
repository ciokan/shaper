package jail

import (
	"github.com/ciokan/shaper/shaper/jail/markers"
	"github.com/ciokan/shaper/shaper/jail/match"
	"github.com/ciokan/shaper/shaper/jail/penalty"
)

const (
	BandwidthJail = iota
	ConnectionsJail
)

type Jail interface {
	Type() int
	GetMatch() match.Match
	GetPenalty() penalty.Penalty
	GetInterface() string
	GetMarker() (found bool, marker *markers.Marker)
	IptablesCmds(delMode bool, marker *markers.Marker) []string
}
