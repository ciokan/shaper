package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	
	"github.com/spf13/cobra"
	
	"github.com/ciokan/shaper/shaper"
)

const (
	CfgFile    = "shaper.yaml"
	ScriptFile = "/tmp/shaper.sh"
)

var (
	err     error
	db      *database
	rootCmd = &cobra.Command{
		Use:   "",
		Short: GCmdRootShort,
		Long:  GCmdRootLong,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("root", args)
		},
	}
	
	applyCmd = &cobra.Command{
		Use:   "apply",
		Short: GCmdApplyShort,
		Long:  GCmdApplyLong,
		Run: func(cmd *cobra.Command, args []string) {
			apply(false)
		},
	}
	
	resetCmd = &cobra.Command{
		Use:   "reset",
		Short: GCmdResetShort,
		Long:  GCmdResetLong,
		Run: func(cmd *cobra.Command, args []string) {
			apply(false)
		},
	}
	
	inspectCmd = &cobra.Command{
		Use:   "inspect",
		Short: GCmdInspectShort,
		Long:  GCmdInspectLong,
		Run: func(cmd *cobra.Command, args []string) {
			s := shaper.New()
			for _, j := range db.Jails {
				jJail, err := j.toJailObj()
				checkErr(err)
				checkErr(s.AddJail(jJail))
			}
			cfg, err := s.Config(false)
			checkErr(err)
			fmt.Println(cfg)
		},
	}
)

func apply(delMode bool) {
	sudo, err := isSudo()
	checkErr(err)
	if sudo == false {
		checkErr(errors.New("this command requires sudo privileges"))
	}
	s := shaper.New()
	for i, j := range db.Jails {
		db.Jails[i].Applied = true
		jJail, err := j.toJailObj()
		checkErr(err)
		checkErr(s.AddJail(jJail))
	}
	checkErr(db.persist())
	cfg, err := s.Config(delMode)
	checkErr(err)
	checkErr(ioutil.WriteFile(ScriptFile, []byte(cfg), 0))
	checkErr(os.Chmod(ScriptFile, 0700))
	defer checkErr(os.Remove(ScriptFile))
	c := exec.Command("sh", ScriptFile)
	out, err := c.CombinedOutput()
	checkErr(err)
	fmt.Println(strings.TrimSpace(string(out)))
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(func() {
		db, err = loadDatabase()
		checkErr(err)
	})
	
	rootCmd.AddCommand(jailCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(resetCmd)
}

func checkErr(err error) {
	if err != nil {
		er(err)
	}
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func isSudo() (bool, error) {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	// 0 = root, 501 = non-root user
	i, err := strconv.Atoi(string(output[:len(output)-1]))
	if err != nil {
		return false, err
	}
	return i == 0, nil
}
