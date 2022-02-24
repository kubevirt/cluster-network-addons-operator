package kubectl

import (
	"bytes"
	"os"
	"os/exec"
)

func Kubectl(command ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(os.Getenv("KUBECTL"), command...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
