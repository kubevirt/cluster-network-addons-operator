package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTLSScannerInstallation(t *testing.T) {
	t.Parallel()
	installer, err := filepath.Abs("../../../hack/install-tls-compliance-operator.sh")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "cluster"), 0755); err != nil {
		t.Fatal(err)
	}
	mock := `#!/usr/bin/env bash
set -euo pipefail
echo "$*" >> "$TEST_DIRECTORY/commands"
case "$1" in
create)
  echo '{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"tls-compliance-operator-system"}}'
  echo '{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"tls-compliance-operator-controller-manager"},"spec":{"replicas":1,"template":{"spec":{"containers":[{"name":"manager","args":["--leader-elect"]}]}}}}'
  ;;
apply) cat > "$TEST_DIRECTORY/applied.json" ;;
esac
`
	if err := os.WriteFile(filepath.Join(directory, "cluster", "kubectl.sh"), []byte(mock), 0755); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("bash", installer)
	command.Dir = directory
	command.Env = append(os.Environ(), "TEST_DIRECTORY="+directory, "VERSION=v1.0.18")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("installer failed: %v\n%s", err, output)
	}
	manifest, err := os.ReadFile(filepath.Join(directory, "applied.json"))
	if err != nil {
		t.Fatal(err)
	}
	var resources struct {
		Items []struct {
			Kind string
			Spec struct {
				Replicas int
				Template struct {
					Spec struct {
						Containers []struct{ Args []string }
					}
				}
			}
		}
	}
	if err := json.Unmarshal(manifest, &resources); err != nil {
		t.Fatal(err)
	}
	if len(resources.Items) != 2 || resources.Items[0].Kind != "Namespace" || resources.Items[1].Spec.Replicas != 0 {
		t.Fatalf("installation must preserve resources and initially stop the scanner: %s", manifest)
	}
	arguments := strings.Join(resources.Items[1].Spec.Template.Spec.Containers[0].Args, " ")
	for _, argument := range []string{"--leader-elect", "--include-namespaces=cluster-network-addons", "--scan-interval=5s", "--cleanup-interval=7s", "--max-retries=0"} {
		if !strings.Contains(arguments, argument) {
			t.Fatalf("initial deployment missing %s: %s", argument, manifest)
		}
	}
	commands, err := os.ReadFile(filepath.Join(directory, "commands"))
	if err != nil {
		t.Fatal(err)
	}
	previousPosition := -1
	for _, step := range []string{"apply -f -", "wait --for=delete pods", "wait --for=condition=Established", "delete tlscompliancereports --all", "delete networkpolicy", "scale deployment", "rollout status"} {
		position := strings.Index(string(commands), step)
		if position <= previousPosition {
			t.Fatalf("missing or unordered step %s:\n%s", step, commands)
		}
		previousPosition = position
	}
}
