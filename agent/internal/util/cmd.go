package util

import (
	"bytes"
	"os/exec"
)

func RunCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

func CmdExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
