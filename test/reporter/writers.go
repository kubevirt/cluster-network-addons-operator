package reporter

import (
	"fmt"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	"os"
	"strings"
)

// Cleanup cleans up the current content of the artifactsDir
func (r *KubernetesCNAOReporter) Cleanup() {
	// clean up artifacts from previous run
	if r.artifactsDir != "" {
		relativeDir := getRootRelativePath(r.artifactsDir)
		if err := os.RemoveAll(relativeDir); err != nil {
			panic(err)
		}
		if err := os.MkdirAll(relativeDir, 0755); err != nil {
			panic(fmt.Sprintf("Error creating directory: %v", err))
		}
	}
}

func (r *KubernetesCNAOReporter) logCommand(args []string, topic string) {
	stdout, stderr, err := kubectl.Kubectl(args...)
	if err != nil {
		fmt.Printf("Error running command kubectl %v, err %v, stderr %s\n", args, err, stderr)
		return
	}

	fileName := fmt.Sprintf(r.artifactsDir+"%d_%s.log", r.failureCount, topic)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error running command %v, err %v\n", args, err)
		return
	}
	defer file.Close()

	if _, err = fmt.Fprint(file, fmt.Sprintf("kubectl %s\n%s\n", strings.Join(args, " "), stdout)); err != nil {
		fmt.Printf("Error writing log %s to file, err %v\n", fileName, err)
	}
}

func (r *KubernetesCNAOReporter) logNamespacePods() {
	args := []string{"get", "pods", "-n", r.namespace, "--no-headers", "-o=custom-columns=NAME:.metadata.name"}
	cnaoPods, stderr, err := kubectl.Kubectl(args...)
	if err != nil {
		fmt.Printf("Error running command kubectl %v, stderr %s, err %v\n", args, stderr, err)
		return
	}

	for _, podName := range strings.Split(cnaoPods, "\n") {
		if podName == "" {
			continue
		}
		args = []string{"get", "pod", podName, "-n", r.namespace, "--no-headers", "-o=jsonpath='{.spec.containers[*].name}'"}

		podContainers, stderr, err := kubectl.Kubectl(args...)
		if err != nil {
			fmt.Printf("Error running command kubectl %v, stderr %s, err %v\n", args, stderr, err)
			return
		}
		podContainers = strings.Trim(podContainers, "'")
		for _, containerName := range strings.Split(podContainers, " ") {
			args = []string{"logs", "-n", r.namespace, podName, "-c", containerName}
			r.logCommand(args, podName+"_"+containerName)

			argsPrev := append(args, "-p")
			if _, _, err := kubectl.Kubectl(argsPrev...); err != nil {
				r.logCommand(argsPrev, "prev_"+podName+"_"+containerName)
			}
		}
	}
}

func getRootRelativePath(artifactsDir string) string {
	rootDepth := strings.Count(artifactsDir, "/")
	for i := 0; i < rootDepth; i++ {
		artifactsDir = "../" + artifactsDir
	}
	return artifactsDir
}
