package penalty

import (
	"bytes"
	"text/template"
)

const (
	DropType = iota
	BandwidthType
)

const (
	TcBwClassCmd  = "$TC class add dev {{.Iface}} parent {{.ParentId}}:1 classid {{.ParentId}}:{{.ClassId}} htb rate {{.Rate}}mbit ceil {{.Ceil}}mbit"
	TcBwQdiskCmd  = "$TC qdisc add dev {{.Iface}} parent {{.ParentId}}:{{.ClassId}} sfq"
	TcBwFilterCmd = "$TC filter add dev {{.Iface}} parent {{.ParentId}}:0 protocol ip prio 1 handle {{.Marker}} fw flowid {{.ParentId}}:{{.ClassId}}"
)

type Penalty interface {
	Type() int
}

// drops connections
type Drop struct{}

func (d Drop) Type() int {
	return DropType
}

// limits bandwidth to specified constraints
type Bandwidth struct {
	Rate uint
	Ceil uint
}

func (b Bandwidth) Type() int {
	return BandwidthType
}

type BwTcCmdParams struct {
	Iface      string
	ClassId    int
	ParentId   int
	Marker     int
	Rate, Ceil uint
}

// generates the required TC commands
func (b Bandwidth) TcCommands(p BwTcCmdParams) ([]string, []string, []string, error) {
	p.Rate = b.Rate
	p.Ceil = b.Ceil
	if p.Ceil == 0 {
		p.Ceil = p.Rate
	}
	
	classCmd, err := b.TcCmd(TcBwClassCmd, p)
	if err != nil {
		return nil, nil, nil, err
	}
	qdiskCmd, err := b.TcCmd(TcBwQdiskCmd, p)
	if err != nil {
		return nil, nil, nil, err
	}
	filterCmd, err := b.TcCmd(TcBwFilterCmd, p)
	if err != nil {
		return nil, nil, nil, err
	}
	return []string{classCmd}, []string{qdiskCmd}, []string{filterCmd}, nil
}

func (b Bandwidth) TcCmd(cmdTemplate string, cmdParams BwTcCmdParams) (string, error) {
	t := template.Must(template.New("tcCmd").Parse(cmdTemplate))
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, cmdParams); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
