package kubectl

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/gomega"
)

func Kubectl(command ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(os.Getenv("KUBECTL"), command...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String() + stderr.String(), err
}

func StartPortForwardCommand(namespace, serviceName string, sourcePort, targetPort int) (*exec.Cmd, error) {
	cmd := exec.Command(os.Getenv("KUBECTL"), "port-forward", "-n", namespace, fmt.Sprintf("service/%s", serviceName), fmt.Sprintf("%d:%d", sourcePort, targetPort))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	waitForPortForwardCmd(stdout, sourcePort, targetPort)
	return cmd, nil
}

func waitForPortForwardCmd(stdout io.ReadCloser, src, dst int) {
	Eventually(func() string {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		Expect(err).NotTo(HaveOccurred())

		return string(tmp)
	}, 30*time.Second, 1*time.Second).Should(ContainSubstring(fmt.Sprintf("Forwarding from 127.0.0.1:%d -> %d", src, dst)))
}

func KillPortForwardCommand(portForwardCmd *exec.Cmd) error {
	if portForwardCmd == nil {
		return nil
	}

	portForwardCmd.Process.Kill()
	_, err := portForwardCmd.Process.Wait()
	return err
}
