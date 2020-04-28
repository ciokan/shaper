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
type Size struct {
	Interface string
	Match     match.FloorCeil
	Penalty   penalty.Penalty
}

func (s Size) Type() int {
	return ConnectionsJail
}

func (s Size) GetMatch() match.Match {
	return s.Match
}

func (s Size) GetPenalty() penalty.Penalty {
	return s.Penalty
}

func (s Size) GetInterface() string {
	return s.Interface
}

func (s Size) markerIdentifier() string {
	switch p := s.Penalty.(type) {
	case penalty.Bandwidth:
		// currently only bandwidth deals with markers
		return fmt.Sprintf("bw:bw:%s:%d:%d", s.Interface, p.Rate, p.Ceil)
	}
	return ""
}

func (s Size) GetMarker() (bool, *markers.Marker) {
	markersObj := markers.New()
	return markersObj.Get(s.markerIdentifier())
}

func (s Size) IptablesCmds(delMode bool, marker *markers.Marker) []string {
	cmds := []string{"$IPT"}
	if delMode {
		cmds = append(cmds, "-D")
	} else {
		cmds = append(cmds, "-A")
	}
	
	cmds = append(cmds, []string{
		"OUTPUT", "-i", s.GetInterface(), "-p", "tcp", "-m", "connbytes", "--connbytes",
	}...)
	
	if s.Match.Floor != 0 {
		if s.Match.Ceil != 0 {
			cmds = append(cmds, fmt.Sprintf("%d:%d", s.Match.Floor, s.Match.Ceil))
		} else {
			cmds = append(cmds, fmt.Sprintf("%d:", s.Match.Floor))
		}
	} else {
		cmds = append(cmds, fmt.Sprintf(":%d", s.Match.Ceil))
	}
	
	cmds = append(cmds, []string{"--connbytes-dir", "both", "--connbytes-mode", "bytes", "-j"}...)
	
	if s.Penalty.Type() == penalty.DropType {
		cmds = append(cmds, "DROP")
	} else {
		cmds = append(cmds, []string{"MARK", "--set-mark", fmt.Sprintf("%d", marker.Value)}...)
	}
	
	return []string{strings.Join(cmds, " ")}
}
