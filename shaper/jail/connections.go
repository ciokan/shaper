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
type Connections struct {
	Interface string
	Match     match.FloorCeil
	Penalty   penalty.Penalty
}

func (c Connections) Type() int {
	return BandwidthJail
}

func (c Connections) GetMatch() match.Match {
	return c.Match
}

func (c Connections) GetPenalty() penalty.Penalty {
	return c.Penalty
}

func (c Connections) GetInterface() string {
	return c.Interface
}

func (c Connections) markerIdentifier() string {
	switch p := c.Penalty.(type) {
	case penalty.Bandwidth:
		// currently only bandwidth deals with markers
		return fmt.Sprintf("conn:conn:%s:%d:%d", c.Interface, p.Rate, p.Ceil)
	}
	return ""
}

func (c Connections) GetMarker() (bool, *markers.Marker) {
	markersObj := markers.New()
	return markersObj.Get(c.markerIdentifier())
}

// The iptables commands are a mix of match and penalty which makes them
// unique per jail so we must generated them at the jail level and not
// on the match or penalty objects, unlike the TC commands which are generated
// by the penalty object
func (c Connections) IptablesCmds(delMode bool, marker *markers.Marker) []string {
	cmds := []string{"$IPT"}
	if delMode {
		cmds = append(cmds, "-D")
	} else {
		cmds = append(cmds, "-A")
	}
	
	cmds = append(cmds, []string{"OUTPUT", "-i", c.GetInterface(), "-p", "tcp", "-m", "connlimit"}...)
	
	if c.Match.Floor != 0 {
		cmds = append(cmds, []string{"--connlimit-above", fmt.Sprintf("%d", c.Match.Floor)}...)
	}
	if c.Match.Ceil != 0 {
		cmds = append(cmds, []string{"--connlimit-upto", fmt.Sprintf("%d", c.Match.Ceil)}...)
	}
	if c.Penalty.Type() == penalty.DropType {
		cmds = append(cmds, []string{"-j", "DROP"}...)
	} else {
		cmds = append(cmds, []string{"-j", "MARK", "--set-mark", fmt.Sprintf("%d", marker.Value)}...)
	}
	
	return []string{strings.Join(cmds, " ")}
}
