package reporter

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
	r.logNamespacePods()

	return r.failureCount
}
