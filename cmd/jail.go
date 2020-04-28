package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	
	"github.com/spf13/cobra"
	
	"github.com/ciokan/shaper/shaper"
	"github.com/ciokan/shaper/shaper/jail"
	"github.com/ciokan/shaper/shaper/jail/match"
	"github.com/ciokan/shaper/shaper/jail/penalty"
)

type jailProps struct {
	Applied          bool   `yaml:"applied"`
	Identifier       string `yaml:"identifier"`
	Interface        string `yaml:"interface"`
	MatchSize        string `yaml:"match-size"`
	MatchConnections string `yaml:"match-connections"`
	PenaltyDrop      bool   `yaml:"penalty-drop"`
	PenaltyBandwidth string `yaml:"penalty-bandwidth"`
}

// creates an identifier based on the provided params
// it's basically just an md5 cut down to 10 chars
// @TODO: find a less moronic way of generating identifiers
// @TODO: one without possible collisions
func (j *jailProps) genId() {
	j.Identifier = str2md5(fmt.Sprintf("%s:%s:%s:%t:%s",
		j.Interface,
		j.MatchSize,
		j.MatchConnections,
		j.PenaltyDrop,
		j.PenaltyBandwidth,
	))[0:10]
}

// entry point that transforms a jail from a cmd param form into a jail object used by the shaper
func (j *jailProps) toJailObj() (jail.Jail, error) {
	var jPenalty penalty.Penalty
	
	// validations
	if j.PenaltyBandwidth != "" {
		var rate, ceil uint64
		rateCeil := strings.Split(j.PenaltyBandwidth, ":")
		
		if rateCeil[0] != "" {
			rate, err = strconv.ParseUint(rateCeil[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("I was unable to convert penalty rate value: %v", err)
			}
		}
		
		if rateCeil[1] != "" {
			ceil, err = strconv.ParseUint(rateCeil[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("I was unable to convert penalty ceil value: %v", err)
			}
		}
		
		jPenalty = penalty.Bandwidth{
			Rate: uint(rate),
			Ceil: uint(ceil),
		}
	}
	
	if j.PenaltyDrop {
		jPenalty = penalty.Drop{}
	}
	
	var mFloor, mCeil uint64
	floorCeil := strings.Split(j.MatchSize, ":")
	if j.MatchConnections != "" {
		floorCeil = strings.Split(j.MatchConnections, ":")
	}
	
	if len(floorCeil) != 2 {
		return nil, fmt.Errorf(`
	floor and ceil values are required; to ommit a ceil value and
	apply a catch-all rule just place a colon after floor. Ex: 1000:`)
	}
	
	if j.MatchSize != "" || j.MatchConnections != "" {
		if floorCeil[0] != "" {
			mFloor, err = strconv.ParseUint(floorCeil[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("I was unable to convert match floor value: %v", err)
			}
		}
		
		if floorCeil[1] != "" {
			mCeil, err = strconv.ParseUint(floorCeil[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("I was unable to convert match ceil value: %v", err)
			}
		}
	}
	
	// now construct the appropriate jail
	if j.MatchSize != "" {
		return jail.Size{
			Interface: j.Interface,
			Match: match.FloorCeil{
				Floor: uint(mFloor),
				Ceil:  uint(mCeil),
			},
			Penalty: jPenalty,
		}, nil
	}
	
	if j.MatchConnections != "" {
		return jail.Connections{
			Interface: j.Interface,
			Match: match.FloorCeil{
				Floor: uint(mFloor),
				Ceil:  uint(mCeil),
			},
			Penalty: jPenalty,
		}, nil
	}
	
	return nil, errors.New("invalid jail definition, make sure there's a match and penalty")
}

var (
	newJail = &jailProps{}
	
	jailCmd = &cobra.Command{
		Use:   "jail",
		Short: GCmdJailShort,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("please call a subcommand")
		},
	}
	
	addJailCmd = &cobra.Command{
		Use:   "add",
		Short: GCmdJailAddShort,
		Long:  GCmdJailAddLong,
		Run: func(cmd *cobra.Command, args []string) {
			// validation first
			jJail, err := newJail.toJailObj()
			checkErr(err)
			
			s := shaper.New()
			checkErr(s.AddJail(jJail))
			
			newJail.genId()
			for _, ex := range db.Jails {
				if ex.Identifier == newJail.Identifier {
					er(errors.New("jail with same params already exists"))
				}
			}
			db.Jails = append(db.Jails, newJail)
			checkErr(db.persist())
		},
	}
	
	delJailCmd = &cobra.Command{
		Use:   "del",
		Short: GCmdJailDelShort,
		Run: func(cmd *cobra.Command, args []string) {
			var newJails []*jailProps
			for _, ex := range db.Jails {
				if ex.Identifier != newJail.Identifier {
					newJails = append(newJails, ex)
				}
			}
			db.Jails = newJails
			checkErr(db.persist())
		},
	}
	
	listJailsCmd = &cobra.Command{
		Use:   "list",
		Short: GCmdJailsListShort,
		Run: func(cmd *cobra.Command, args []string) {
			jStr, err := db.jailsYaml()
			checkErr(err)
			fmt.Println(jStr)
		},
	}
)

func addInit() {
	jailCmd.AddCommand(addJailCmd)
	// extract main interface to use it as a default
	mainIf, err := mainInterface()
	if err != nil {
		er(err)
	}
	addJailCmd.Flags().StringVarP(&newJail.Interface, "interface", "i", mainIf, GInterface)
	addJailCmd.Flags().StringVar(&newJail.MatchSize, "match-size", "", GMatchSize)
	addJailCmd.Flags().StringVar(&newJail.MatchConnections, "match-connections", "", GMatchConnections)
	addJailCmd.Flags().StringVar(&newJail.PenaltyBandwidth, "penalty-bandwidth", "", GPenaltyBandwidth)
	addJailCmd.Flags().BoolVar(&newJail.PenaltyDrop, "penalty-drop", false, GPenaltyDrop)
}

func delInit() {
	jailCmd.AddCommand(delJailCmd)
	delJailCmd.Flags().StringVar(&newJail.Identifier, "identifier", "", GIdentifier)
}

func listInit() {
	jailCmd.AddCommand(listJailsCmd)
}

func init() {
	addInit()
	delInit()
	listInit()
}
