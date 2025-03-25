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
	r.logNamespacePods()

	return r.failureCount
}
