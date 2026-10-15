package test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var tlsScannerInstallerPath, tlsScannerInstallerPathError = filepath.Abs("../../../hack/install-tls-compliance-operator.sh")

func TestTLSScannerInstallation(t *testing.T) {
	t.Parallel()
	if tlsScannerInstallerPathError != nil {
		t.Fatal(tlsScannerInstallerPathError)
	}
	for _, scenario := range []struct {
		name       string
		reports    string
		failure    string
		diagnostic string
	}{
		{name: "scoped reports", reports: `{"items":[{"metadata":{"name":"scoped-report"},"spec":{"sourceNamespace":"cluster-network-addons"}}]}`},
		{name: "no reports", reports: `{"items":[]}`},
		{name: "unrelated report", reports: `{"items":[{"metadata":{"name":"scoped-report"},"spec":{"sourceNamespace":"cluster-network-addons"}},{"metadata":{"name":"unrelated-report"},"spec":{"sourceNamespace":"other-namespace"}}]}`, diagnostic: "existing reports outside cluster-network-addons"},
		{name: "pod wait failure", reports: `{"items":[]}`, failure: "pods", diagnostic: "injected pods failure"},
		{name: "create failure", reports: `{"items":[]}`, failure: "create", diagnostic: "injected create failure"},
		{name: "download failure", reports: `{"items":[]}`, failure: "curl", diagnostic: "injected curl failure"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for _, name := range []string{"cluster", "bin"} {
				if err := os.Mkdir(filepath.Join(directory, name), 0755); err != nil {
					t.Fatal(err)
				}
			}
			originalResources := []json.RawMessage{
				json.RawMessage(`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"tls-compliance-operator-system","labels":{"preserve":"true"}}}`),
				json.RawMessage(`{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"tls-compliance-operator-controller-manager"},"spec":{"replicas":1,"template":{"spec":{"containers":[{"name":"manager","image":"scanner:fixture","args":["--leader-elect","--metrics-bind-address=:8443"]},{"name":"sidecar","args":["--unchanged"]}]}}}}`),
				json.RawMessage(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"preserved-config"},"data":{"key":"value"}}`),
			}
			manifest := make([]string, len(originalResources))
			for index, resource := range originalResources {
				manifest[index] = string(resource)
			}
			files := map[string]string{
				"resources.json": strings.Join(manifest, "\n"),
				"reports.json":   scenario.reports,
				"bin/curl": `#!/usr/bin/env bash
set -euo pipefail
echo "curl $*" >> "$TEST_DIRECTORY/commands"
if [[ "$TEST_FAILURE" == curl ]]; then
  echo "injected curl failure" >&2
  exit 23
fi
output=""
while (($#)); do
  case "$1" in
    -o|--output) output="$2"; shift ;;
    --output=*) output="${1#*=}" ;;
  esac
  shift
done
test -n "$output"
cp "$TEST_DIRECTORY/resources.json" "$output"
`,
				"cluster/kubectl.sh": `#!/usr/bin/env bash
set -euo pipefail
operation=""
for argument in "$@"; do
  case "$argument" in
    create|apply|rollout|wait|get|delete|scale) operation="$argument"; break ;;
  esac
done
if [[ "$operation" == apply ]]; then
  cat > "$TEST_DIRECTORY/applied.json"
fi
echo "$*" >> "$TEST_DIRECTORY/commands"
case "$operation" in
  create)
    if [[ "$TEST_FAILURE" == create ]]; then
      echo "injected create failure" >&2
      exit 23
    fi
    input=""
    while (($#)); do
      case "$1" in
        -f|--filename) input="$2"; shift ;;
        --filename=*) input="${1#*=}" ;;
      esac
      shift
    done
    test -n "$input"
    cat "$input"
    ;;
  wait)
    if [[ "$TEST_FAILURE" == pods && " $* " == *" pods "* ]]; then
      echo "injected pods failure" >&2
      exit 23
    fi
    ;;
  get) cat "$TEST_REPORTS_FILE" ;;
esac
`,
			}
			for name, contents := range files {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0755); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "bash", tlsScannerInstallerPath)
			command.WaitDelay = time.Second
			command.Dir = directory
			command.Env = append(os.Environ(), "TEST_DIRECTORY="+directory, "VERSION=v1.0.18",
				"TEST_FAILURE="+scenario.failure, "TEST_REPORTS_FILE="+filepath.Join(directory, "reports.json"),
				"PATH="+filepath.Join(directory, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
			output, commandError := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("installer exceeded command deadline: %v\n%s", ctx.Err(), output)
			}
			if scenario.diagnostic == "" && commandError != nil {
				t.Fatalf("installer failed: %v\n%s", commandError, output)
			}
			if scenario.diagnostic != "" && (commandError == nil || !strings.Contains(string(output), scenario.diagnostic)) {
				t.Fatalf("expected failure containing %q, got %v:\n%s", scenario.diagnostic, commandError, output)
			}
			log, err := os.ReadFile(filepath.Join(directory, "commands"))
			if err != nil {
				t.Fatal(err)
			}
			commands := strings.Split(strings.TrimSpace(string(log)), "\n")
			for _, entry := range commands {
				if strings.Contains(entry, "delete") && strings.Contains(entry, "tlscompliancereport") {
					t.Fatalf("installer must never delete reports:\n%s", log)
				}
				if scenario.diagnostic != "" && strings.Contains(entry, "scale") {
					t.Fatalf("installer must not start the scanner after failure:\n%s", log)
				}
			}
			assertTLSScannerCommand(t, commands[0], "curl", "--fail", "--silent", "--show-error", "--location", "--max-time=60",
				"https://github.com/sebrandon1/tls-compliance-operator/releases/download/v1.0.18/install.yaml", "-o")
			if scenario.failure == "curl" || scenario.failure == "create" {
				return
			}
			applied, err := os.ReadFile(filepath.Join(directory, "applied.json"))
			if err != nil {
				t.Fatal(err)
			}
			var resources struct {
				APIVersion string
				Kind       string
				Items      []json.RawMessage
			}
			if err := json.Unmarshal(applied, &resources); err != nil {
				t.Fatal(err)
			}
			if resources.APIVersion != "v1" || resources.Kind != "List" || len(resources.Items) != len(originalResources) {
				t.Fatalf("installation must apply a List preserving every resource: %s", applied)
			}
			for index, original := range originalResources {
				var expected, actual map[string]interface{}
				if err := json.Unmarshal(original, &expected); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(resources.Items[index], &actual); err != nil {
					t.Fatal(err)
				}
				if index == 1 {
					spec := expected["spec"].(map[string]interface{})
					spec["replicas"] = float64(0)
					containers := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
					manager := containers[0].(map[string]interface{})
					manager["args"] = append(manager["args"].([]interface{}), "--include-namespaces=cluster-network-addons",
						"--scan-interval=5s", "--cleanup-interval=7s", "--max-retries=0")
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("unexpected manifest modification for resource %d:\nexpected: %v\nactual: %v", index, expected, actual)
				}
			}
			steps := [][]string{
				{"create", "--dry-run=client", "-f", "-o=json"},
				{"apply", "-f", "-"},
				{"rollout", "status", "deployment/tls-compliance-operator-controller-manager", "-n=tls-compliance-operator-system", "--timeout=120s", "--request-timeout=15s"},
				{"wait", "--for=delete", "pods", "-n=tls-compliance-operator-system", "-l=app.kubernetes.io/name=tls-compliance-operator,control-plane=controller-manager", "--timeout=120s"},
				{"wait", "--for=condition=Established", "crd/tlscompliancereports.security.telco.openshift.io", "--timeout=120s"},
				{"get", "tlscompliancereports", "-o=json"},
				{"delete", "networkpolicy", "tls-compliance-operator-controller-manager", "-n=tls-compliance-operator-system"},
				{"scale", "deployment", "tls-compliance-operator-controller-manager", "--replicas=1", "-n=tls-compliance-operator-system"},
				{"rollout", "status", "deployment/tls-compliance-operator-controller-manager", "-n=tls-compliance-operator-system", "--timeout=120s"},
			}
			if scenario.failure == "pods" {
				steps = steps[:4]
			} else if scenario.diagnostic != "" {
				steps = steps[:6]
			}
			if len(commands) != len(steps)+1 {
				t.Fatalf("unexpected installer commands:\n%s", log)
			}
			for index, step := range steps {
				assertTLSScannerCommand(t, commands[index+1], step...)
			}
			createArguments := strings.Fields(commands[1])
			for index, argument := range createArguments {
				if argument == "-f" && index+1 < len(createArguments) {
					if strings.Contains(createArguments[index+1], "://") {
						t.Fatal("kubectl create must consume the local download, not fetch the URL")
					}
				}
			}
		})
	}
}

func assertTLSScannerCommand(t *testing.T, command string, expected ...string) {
	t.Helper()
	arguments := strings.Fields(command)
	for _, argument := range expected {
		found := false
		for index, actual := range arguments {
			if actual == argument {
				found = true
				break
			}
			if flag, value, hasValue := strings.Cut(argument, "="); hasValue && actual == flag && index+1 < len(arguments) && arguments[index+1] == value {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("command %q missing argument %q", command, argument)
		}
	}
}
