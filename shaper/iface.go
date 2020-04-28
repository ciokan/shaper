package shaper

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	
	"github.com/ciokan/shaper/shaper/jail"
	"github.com/ciokan/shaper/shaper/jail/match"
	"github.com/ciokan/shaper/shaper/jail/penalty"
)

type Iface struct {
	name            string
	jails           []jail.Jail
	tcClassId       int // parent classId (group id?) for TC commands
	tcClassParentId int // classId for TC commands
}

type TcCommands struct {
	Class  []string
	Disk   []string
	Filter []string
}

func NewIface(name string, tcClassParentId int) *Iface {
	return &Iface{
		name:            name,
		tcClassId:       3,
		tcClassParentId: tcClassParentId,
	}
}

func (i *Iface) tcCid() int {
	// Returns the next available class id for our tc groups
	// Right now it operates as a smple incremented counter basically
	cId := i.tcClassId
	i.tcClassId += 1
	return cId
}

// returns jails with same match type
func (i *Iface) sameTypeJails(j jail.Jail) []jail.Jail {
	var same []jail.Jail
	for _, existing := range i.jails {
		if existing.GetMatch().Type() == j.GetMatch().Type() && existing.Type() == j.Type() {
			same = append(same, existing)
		}
	}
	return same
}

func (i *Iface) addJail(j jail.Jail) error {
	if err := i.validateJail(j); err != nil {
		return err
	}
	i.jails = append(i.jails, j)
	return nil
}

func (i *Iface) validateJail(j jail.Jail) error {
	if j.GetMatch().Type() == match.FloorCeilType {
		return i.validateMatchFloorCeil(j)
	}
	return nil
}

// makes sure that the values do not overlap
func (i *Iface) validateMatchFloorCeil(j jail.Jail) error {
	var maxCeil uint
	newMatch := j.GetMatch().(match.FloorCeil)
	
	for _, existing := range i.sameTypeJails(j) {
		exMatch := existing.GetMatch().(match.FloorCeil)
		
		if exMatch.Ceil == 0 && newMatch.Ceil >= exMatch.Floor {
			return fmt.Errorf("a catch-all match was found - the new match overlaps it")
		}
		if exMatch.Ceil == newMatch.Ceil && exMatch.Floor == newMatch.Floor {
			return fmt.Errorf("a match with the same parameters was already added")
		}
		if newMatch.Overlaps(exMatch) {
			return fmt.Errorf(
				"match rule for %s is overlapping an existing one. Floor: %d, Ceil: %d",
				i.name, exMatch.Floor, exMatch.Ceil,
			)
		}
		// store max exit value because we need it later
		if exMatch.Ceil > maxCeil {
			maxCeil = exMatch.Ceil
		}
	}
	// if this new rule has no exit barrier it means it's a catch-all that
	// applies to everything that is "bigger than the rest" which also means
	// that its entry value must be bigger than all the existing exit values
	// currently in store
	if newMatch.Ceil == 0 && newMatch.Floor <= maxCeil {
		return fmt.Errorf("the new catch-all rule has a floor value lower than the ceil of another")
	}
	return nil
}

func (i *Iface) TcCmds(j jail.Jail) (*TcCommands, error) {
	switch p := j.GetPenalty().(type) {
	case penalty.Bandwidth:
		// we need markers for this penalty
		found, marker := j.GetMarker()
		if found {
			// not an error
			return nil, nil
		}
		classId := i.tcCid()
		tcClass, tcQdisk, tcFilter, err := p.TcCommands(penalty.BwTcCmdParams{
			Iface:    j.GetInterface(),
			ClassId:  classId,
			ParentId: i.tcClassParentId,
			Marker:   marker.Value,
		})
		if err != nil {
			return nil, err
		}
		return &TcCommands{
			Class:  tcClass,
			Disk:   tcQdisk,
			Filter: tcFilter,
		}, nil
	case penalty.Drop:
		// drop penalty needs no TC commands (drops straight from iptables)
	}
	return nil, nil
}

// creates the final script for this interfaces and its jails
func (i *Iface) Script(delMode bool) (string, error) {
	type scriptParams struct {
		TcClass         string
		TcQdisk         string
		TcFilter        string
		Iptables        string
		Interface       string
		TcClassParentId int
		DelMode         bool
	}
	
	var tcClassCmds, tcQdiskCmds, tcFilterCmds, iptablesCmds []string
	for _, j := range i.jails {
		tcCmds, err := i.TcCmds(j)
		if err != nil {
			return "", err
		}
		if tcCmds != nil && delMode != true {
			tcClassCmds = append(tcClassCmds, tcCmds.Class...)
			tcQdiskCmds = append(tcQdiskCmds, tcCmds.Disk...)
			tcFilterCmds = append(tcFilterCmds, tcCmds.Filter...)
		}
		_, marker := j.GetMarker()
		iptablesCmds = append(iptablesCmds, j.IptablesCmds(delMode, marker)...)
	}
	
	params := scriptParams{
		Interface:       i.name,
		TcClassParentId: i.tcClassParentId,
		TcClass:         strings.Join(tcClassCmds, "\n"),
		TcQdisk:         strings.Join(tcQdiskCmds, "\n"),
		TcFilter:        strings.Join(tcFilterCmds, "\n"),
		Iptables:        strings.Join(iptablesCmds, "\n"),
		DelMode:         delMode,
	}
	t := template.Must(template.New("ifaceScript").Parse(IfaceScriptTemplate))
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, params); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// @TODO: get the total and desired bw from user or find a way to determine it (10000mbit hint)
const IfaceScriptTemplate = `
# {{.Interface}}
# ---------------------------------------------------------
$TC qdisc del dev {{.Interface}} root

{{ if .DelMode }}
$IPT -D PREROUTING -i {{.Interface}} -t mangle -j CONNMARK --restore-mark
$IPT -D POSTROUTING -i {{.Interface}} -t mangle -m mark ! --mark 0 -j ACCEPT
$IPT -D POSTROUTING -i {{.Interface}} -t mangle -j CONNMARK --save-mark
{{ else }}
$TC qdisc add dev {{.Interface}} root handle {{.TcClassParentId}}: htb default 2

# total and desired bandwidths: WIP
$TC class add dev {{.Interface}} parent {{.TcClassParentId}}: classid {{.TcClassParentId}}:1 htb rate 10000mbit ceil 10000mbit
$TC class add dev {{.Interface}} parent {{.TcClassParentId}}:1 classid {{.TcClassParentId}}:2 htb rate 10000mbit ceil 10000mbit
{{.TcClass}}

$TC qdisc add dev {{.Interface}} parent {{.TcClassParentId}}:2 sfq
{{.TcQdisk}}

{{.TcFilter}}

$IPT -A PREROUTING -i {{.Interface}} -t mangle -j CONNMARK --restore-mark
$IPT -A POSTROUTING -i {{.Interface}} -t mangle -m mark ! --mark 0 -j ACCEPT
$IPT -A POSTROUTING -i {{.Interface}} -t mangle -j CONNMARK --save-mark
{{ end }}

{{.Iptables}}
`
