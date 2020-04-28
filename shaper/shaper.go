package shaper

import (
	"fmt"
	"os/exec"
	"strings"
	
	"github.com/ciokan/shaper/shaper/jail"
)

type JailCmds struct {
}

type Shaper struct {
	ifaces []*Iface
}

var shaperObj *Shaper

func New() *Shaper {
	if shaperObj == nil {
		shaperObj = &Shaper{}
	}
	return shaperObj
}

func (s *Shaper) AddJail(j jail.Jail) error {
	iface := s.ifaceByName(j.GetInterface())
	if iface == nil {
		iface = NewIface(j.GetInterface(), len(s.ifaces)+1)
		s.ifaces = append(s.ifaces, iface)
	}
	return iface.addJail(j)
}

// returns true if an interface exists in our store (there's a jail with it already)
func (s *Shaper) ifaceByName(iface string) *Iface {
	for _, exIface := range s.ifaces {
		if exIface.name == iface {
			return exIface
		}
	}
	return nil
}

type TplParams struct {
	Tc         string
	Ipt        string
	Interface  string
	TcClass    string
	TcQdisk    string
	TcFilter   string
	IpTbOutput string
}

func (s *Shaper) Config(delMode bool) (string, error) {
	iptExe, err := which("iptables")
	if err != nil {
		return "", fmt.Errorf("error fetching iptables path: %v", err)
	}
	if iptExe == "" {
		return "", fmt.Errorf("iptables executable was not found in path: %v", err)
	}
	
	tcExe, err := which("tc")
	if err != nil {
		return "", fmt.Errorf("error fetching tc path path: %v", err)
	}
	if tcExe == "" {
		return "", fmt.Errorf("iptables executable was not found in path: %v", err)
	}
	
	var ifacesScripts []string
	for _, iface := range s.ifaces {
		script, err := iface.Script(delMode)
		if err != nil {
			return "", fmt.Errorf(
				"error generating script for interface(%s): %v", iface.name, err)
		}
		ifacesScripts = append(ifacesScripts, script)
	}
	
	return fmt.Sprintf(ScriptTemplate, tcExe, iptExe, strings.Join(ifacesScripts, "\n")), nil
}

func which(executable string) (string, error) {
	cmd := exec.Command("which", executable)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

const ScriptTemplate = `#!/bin/sh
$TC=%s
$IPT=%s

%s`
