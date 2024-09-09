package main

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/monitoring/rules"
)

const promImage = "quay.io/prometheus/prometheus:v2.15.2"

func main() {
	if len(os.Args) != 4 {
		panic("Usage: verify-rules <oci-bin> <target-file> <tests-file>")
	}
	ociBin := os.Args[1]
	targetFile := os.Args[2]
	rulesFile := os.Args[3]

	err := verify(ociBin, targetFile, rulesFile)
	if err != nil {
		panic(err)
	}
}

func verify(ociBin string, targetFile string, rulesFile string) error {
	defer deleteRulesFile(targetFile)
	err := createRulesFile(targetFile)
	if err != nil {
		return err
	}

	err = lint(ociBin, targetFile)
	if err != nil {
		return err
	}

	err = unitTest(ociBin, targetFile, rulesFile)
	if err != nil {
		return err
	}

	return nil
}

func createRulesFile(targetFile string) error {
	if err := rules.SetupRules("ci"); err != nil {
		return err
	}

	promRule, err := rules.BuildPrometheusRule("ci")
	if err != nil {
		return err
	}

	b, err := json.Marshal(promRule.Spec)
	if err != nil {
		return err
	}

	err = os.WriteFile(targetFile, b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func deleteRulesFile(targetFile string) error {
	err := os.Remove(targetFile)
	if err != nil {
		return err
	}

	return nil
}

func lint(ociBin string, targetFile string) error {
	cmd := exec.Command(ociBin, "run", "--rm", "--entrypoint=/bin/promtool", "-v", targetFile+":/tmp/rules.verify:ro,Z", promImage, "check", "rules", "/tmp/rules.verify")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func unitTest(ociBin string, targetFile string, testsFile string) error {
	cmd := exec.Command(ociBin, "run", "--rm", "--entrypoint=/bin/promtool", "-v", testsFile+":/tmp/rules.test:ro,Z", "-v", targetFile+":/tmp/rules.verify:ro,Z", promImage, "test", "rules", "/tmp/rules.test")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
