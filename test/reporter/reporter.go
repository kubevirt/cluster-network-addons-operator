package reporter

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
)

type KubernetesCNAOReporter struct {
	artifactsDir string
	namespace    string
	failureCount int
}

func New(artifactsDir, namespace string) *KubernetesCNAOReporter {
	return &KubernetesCNAOReporter{
		artifactsDir: artifactsDir,
		namespace:    namespace,
		failureCount: 0,
	}
}

// DumpLogs saves the desired logs to files
func (r *KubernetesCNAOReporter) DumpLogs() int {
	r.failureCount++
	r.logCommand([]string{"get", "all", "-A"}, "overview")
	r.logCommand([]string{"get", "networkaddonsconfig", "cluster", "-n", r.namespace, "-o", "yaml"}, "networkaddonsconfig")
	r.logCommand([]string{"get", "daemonset", "-n", r.namespace, "-o", "yaml"}, "daemonsets")
	r.logCommand([]string{"get", "deployment", "-n", r.namespace, "-o", "yaml"}, "deployments")
	r.logCommand([]string{"get", "pod", "-n", r.namespace, "-o", "yaml"}, "pods")
	r.logCommand([]string{"get", "clusterrole", "-n", r.namespace, "-o", "yaml"}, "clusterroles")
	r.logCommand([]string{"get", "role", "-n", r.namespace, "-o", "yaml"}, "roles")
	r.logCommand([]string{"get", "rolebinding", "-n", r.namespace, "-o", "yaml"}, "rolebindings")
	r.logCommand([]string{"get", "clusterrolebinding", "-n", r.namespace, "-o", "yaml"}, "clusterrolebindings")
	r.logCommand([]string{"get", "servicemonitor", "-n", r.namespace, "-o", "yaml"}, "servicemonitors")
	r.logCommand([]string{"get", "prometheusrule", "-n", r.namespace, "-o", "yaml"}, "prometheusrules")
	r.logCommand([]string{"get", "service", "-n", r.namespace, "-o", "yaml"}, "services")
	r.logCommand([]string{"get", "endpointslice", "-n", r.namespace, "-o", "yaml"}, "endpointslices")
	r.logNamespacePods()

	r.logCommand([]string{"get", "tlscompliancereport"}, "tlscompliancereports-overview")
	r.logCommand([]string{"get", "tlscompliancereport", "-o", "yaml"}, "tlscompliancereports")

	return r.failureCount
}

func (r *KubernetesCNAOReporter) DumpTLSComplianceReports() error {
	//TODO: move upper the stack and reuse
	ociBin := os.Getenv("OCI_BIN")
	tlsReportImage := os.Getenv("TLSREPORT_IMAGE")
	kubeconfig := os.Getenv("KUBECONFIG")
	artifactsDir := os.Getenv("ARTIFACTS")

	reportsDir := artifactsDir + "/tls-compliance"
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return err
	}

	execTlsReportContainer := func(args ...string) []string {
		c := []string{"run", "--rm", "--network", "host",
			"-v", kubeconfig + ":/root/c:Z,ro", "-e", "KUBECONFIG=/root/c",
			tlsReportImage,
			"kubectl", "tlsreport",
		}
		return append(c, args...)
	}

	overview, stderr, err := kubectl.Kubectl("get", "tlscompliancereport")
	if err != nil {
		return fmt.Errorf("failed to get tlscompliancereports: %w, stderr: \n%s\n, stdout: \n%s\n", err, stderr, overview)
	}
	p := reportsDir + "/tlscompliancereports-overview.log"
	if err := os.WriteFile(p, []byte(overview), 0644); err != nil {
		return err
	}

	csvReport, err := exec.Command(ociBin, execTlsReportContainer("csv")...).Output()
	if err != nil {
		return err
	}
	p = reportsDir + "/tlscompliance.csv"
	if err := os.WriteFile(p, csvReport, 0644); err != nil {
		return err
	}

	junitReport, err := exec.Command(ociBin, execTlsReportContainer("junit")...).Output()
	if err != nil {
		return err
	}
	// junit report is saved on artifacts root dir to enable Prow present it on Spyglass page
	p = artifactsDir + "/junit.tlscompliance.xml"
	if err := os.WriteFile(p, junitReport, 0644); err != nil {
		return err
	}

	return nil
}
