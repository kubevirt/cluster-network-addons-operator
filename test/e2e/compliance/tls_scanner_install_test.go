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

type tlsScannerInstallScenario struct {
	name       string
	reports    string
	failure    string
	diagnostic string
}

func TestTLSScannerInstallation(t *testing.T) {
	t.Parallel()
	if tlsScannerInstallerPathError != nil {
		t.Fatal(tlsScannerInstallerPathError)
	}
	for _, scenario := range []tlsScannerInstallScenario{
		{name: "scoped reports", reports: `{"items":[{"metadata":{"name":"scoped-report"},"spec":{"sourceNamespace":"cluster-network-addons"}}]}`},
		{name: "no reports", reports: `{"items":[]}`},
		{name: "unrelated report", reports: `{"items":[{"metadata":{"name":"scoped-report"},"spec":{"sourceNamespace":"cluster-network-addons"}},{"metadata":{"name":"unrelated-report"},"spec":{"sourceNamespace":"other-namespace"}}]}`, diagnostic: "existing reports outside cluster-network-addons"},
		{name: "pod wait failure", reports: `{"items":[]}`, failure: "pods", diagnostic: "injected pods failure"},
		{name: "create failure", reports: `{"items":[]}`, failure: "create", diagnostic: "injected create failure"},
		{name: "download failure", reports: `{"items":[]}`, failure: "curl", diagnostic: "injected curl failure"},
	} {

		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			directory := prepareTLSScannerInstallation(t, scenario)
			runTLSScannerInstallation(t, directory, scenario)
			commands := tlsScannerInstallCommands(t, directory, scenario)
			if !validTLSScannerDownload(commands[0]) {
				t.Fatalf("installer download must use exact HTTPS-only options: %s", commands[0])
			}
			if scenario.failure == "curl" || scenario.failure == "create" {
				return
			}
			assertTLSScannerManifest(t, directory)
			assertTLSScannerInstallSteps(t, commands, scenario)
			assertTLSScannerLocalManifest(t, commands[1])
		})
	}
}

func tlsScannerOriginalResources() []json.RawMessage {
	return []json.RawMessage{
		json.RawMessage(`{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"tls-compliance-operator-system","labels":{"preserve":"true"}}}`),
		json.RawMessage(`{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"tls-compliance-operator-controller-manager"},"spec":{"replicas":1,"template":{"spec":{"containers":[{"name":"manager","image":"scanner:fixture","args":["--leader-elect","--metrics-bind-address=:8443"]},{"name":"sidecar","args":["--unchanged"]}]}}}}`),
		json.RawMessage(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"preserved-config"},"data":{"key":"value"}}`),
	}
}

func prepareTLSScannerInstallation(t *testing.T, scenario tlsScannerInstallScenario) string {
	t.Helper()
	directory := t.TempDir()
	for _, name := range []string{"cluster", "bin"} {
		if err := os.Mkdir(filepath.Join(directory, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	originalResources := tlsScannerOriginalResources()
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
	return directory
}

func runTLSScannerInstallation(t *testing.T, directory string, scenario tlsScannerInstallScenario) {
	t.Helper()
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
}

func tlsScannerInstallCommands(t *testing.T, directory string, scenario tlsScannerInstallScenario) []string {
	t.Helper()
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
	return commands
}

func assertTLSScannerManifest(t *testing.T, directory string) {
	t.Helper()
	originalResources := tlsScannerOriginalResources()
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
		assertTLSScannerResource(t, original, resources.Items[index])
	}
}

func assertTLSScannerResource(t *testing.T, original, resource json.RawMessage) {
	t.Helper()
	var expected, actual map[string]interface{}
	if err := json.Unmarshal(original, &expected); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(resource, &actual); err != nil {
		t.Fatal(err)
	}
	if expected["kind"] == "Deployment" {
		spec := expected["spec"].(map[string]interface{})
		spec["replicas"] = float64(0)
		containers := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]interface{})
		manager := containers[0].(map[string]interface{})
		manager["args"] = append(manager["args"].([]interface{}), "--include-namespaces=cluster-network-addons",
			"--scan-interval=5s", "--cleanup-interval=7s", "--max-retries=0")
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected manifest modification for resource:\nexpected: %v\nactual: %v", expected, actual)
	}
}

func assertTLSScannerInstallSteps(t *testing.T, commands []string, scenario tlsScannerInstallScenario) {
	t.Helper()
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
		t.Fatalf("unexpected installer commands: %v", commands)
	}
	for index, step := range steps {
		assertTLSScannerCommand(t, commands[index+1], step...)
	}
}

func assertTLSScannerLocalManifest(t *testing.T, command string) {
	t.Helper()
	createArguments := strings.Fields(command)
	for index, argument := range createArguments {
		if argument == "-f" && index+1 < len(createArguments) {
			if strings.Contains(createArguments[index+1], "://") {
				t.Fatal("kubectl create must consume the local download, not fetch the URL")
			}
		}
	}
}

func validTLSScannerDownload(command string) bool {
	expected := []string{"curl", "--fail", "--silent", "--show-error", "--location",
		"--proto", "=https", "--proto-redir", "=https", "--max-time", "60",
		"https://github.com/sebrandon1/tls-compliance-operator/releases/download/v1.0.18/install.yaml", "-o"}
	arguments := strings.Fields(command)
	return len(arguments) == len(expected)+1 && reflect.DeepEqual(arguments[:len(expected)], expected)
}

func TestTLSScannerDownloadPolicy(t *testing.T) {
	t.Parallel()
	secure := "curl --fail --silent --show-error --location --proto =https --proto-redir =https --max-time 60 https://github.com/sebrandon1/tls-compliance-operator/releases/download/v1.0.18/install.yaml -o /tmp/manifest"
	for _, scenario := range []struct {
		name    string
		command string
		allowed bool
	}{
		{name: "HTTPS only", command: secure, allowed: true},
		{name: "initial protocol unrestricted", command: strings.Replace(secure, "--proto =https ", "", 1)},
		{name: "redirect protocol unrestricted", command: strings.Replace(secure, "--proto-redir =https ", "", 1)},
		{name: "initial HTTP allowed", command: strings.Replace(secure, "--proto =https", "--proto =http,https", 1)},
		{name: "redirect HTTP allowed", command: strings.Replace(secure, "--proto-redir =https", "--proto-redir =http,https", 1)},
		{name: "additive protocol selection", command: strings.ReplaceAll(secure, "=https", "+https")},
		{name: "later redirect override", command: secure + " --proto-redir =all"},
		{name: "later initial override", command: secure + " --proto =all"},
		{name: "insecure TLS", command: secure + " --insecure"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if validTLSScannerDownload(scenario.command) != scenario.allowed {
				t.Fatalf("unexpected download policy decision for %q", scenario.command)
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
