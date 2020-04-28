package jail

import (
	"fmt"
	"strings"
	
	"github.com/ciokan/shaper/shaper/jail/markers"
	"github.com/ciokan/shaper/shaper/jail/match"
	"github.com/ciokan/shaper/shaper/jail/penalty"
)

// creates a jail that applies based on the number of (TCP) connections
// created by the remote party - once the user goes between the entry
// and exit limits it is jailed and bandwidth limiting goes into effect
// if the exit limit is 0 then the rule will match for everything that
// goes beyonf the entry limit
type Bandwidth struct {
	Interface string
	Match     match.FloorCeil
	Penalty   penalty.Penalty
}

func (b Bandwidth) Type() int {
	return ConnectionsJail
}

func (b Bandwidth) GetMatch() match.Match {
	return b.Match
}

func (b Bandwidth) GetPenalty() penalty.Penalty {
	return b.Penalty
}

func (b Bandwidth) GetInterface() string {
	return b.Interface
}

func (b Bandwidth) markerIdentifier() string {
	switch p := b.Penalty.(type) {
	case penalty.Bandwidth:
		// currently only bandwidth deals with markers
		return fmt.Sprintf("bw:bw:%s:%d:%d", b.Interface, p.Rate, p.Ceil)
	}
	return ""
}

func (b Bandwidth) GetMarker() (bool, *markers.Marker) {
	markersObj := markers.New()
	return markersObj.Get(b.markerIdentifier())
}

func (b Bandwidth) IptablesCmds(delMode bool, marker *markers.Marker) []string {
	cmds := []string{"$IPT"}
	if delMode {
		cmds = append(cmds, "-D")
	} else {
		cmds = append(cmds, "-A")
	}
	
	cmds = append(cmds, []string{
		"OUTPUT", "-i", b.GetInterface(), "-p", "tcp", "-m", "connbytes", "--connbytes",
	}...)
	
	if b.Match.Floor != 0 {
		if b.Match.Ceil != 0 {
			cmds = append(cmds, fmt.Sprintf("%d:%d", b.Match.Floor, b.Match.Ceil))
		} else {
			cmds = append(cmds, fmt.Sprintf("%d:", b.Match.Floor))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf(":%d", b.Match.Ceil))
	}
	
	cmds = append(cmds, []string{"--connbytes-dir", "both", "--connbytes-mode", "bytes", "-j"}...)
	
	if b.Penalty.Type() == penalty.DropType {
		cmds = append(cmds, "DROP")
	} else {
		cmds = append(cmds, []string{"MARK", "--set-mark", fmt.Sprintf("%d", marker.Value)}...)
	}
	
	return []string{strings.Join(cmds, " ")}
}
