package cmd

import (
	"crypto/md5"
	"encoding/hex"
	"os/exec"
	"strings"
)

func execCmd(command string) (string, error) {
	out, err := exec.Command("sh", "-c", command).Output()
	return strings.TrimSpace(string(out)), err
}

func mainInterface() (string, error) {
	cmd, err := execCmd(`route | grep '^default' | grep -o '[^ ]*$'`)
	if err != nil {
		return "", err
	}
	return strings.Split(strings.TrimSpace(cmd), "\n")[0], nil
}

func str2md5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
