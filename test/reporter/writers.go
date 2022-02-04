package reporter

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"

	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
)

func writePodsLogs(writer io.Writer, namespace string, sinceTime time.Time) {
	podLogOpts := corev1.PodLogOptions{}
	podLogOpts.SinceTime = &metav1.Time{Time: sinceTime}
	podList := &corev1.PodList{}
	err := testenv.Client.List(context.TODO(), podList, &dynclient.ListOptions{Namespace: namespace})
	Expect(err).ToNot(HaveOccurred())
	podsClientset := testenv.KubeClient.CoreV1().Pods(namespace)

	for _, pod := range podList.Items {
		req := podsClientset.GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			io.WriteString(GinkgoWriter, fmt.Sprintf("error in opening stream: %v\n", err))
			continue
		}
		defer podLogs.Close()
		rawLogs, err := ioutil.ReadAll(podLogs)
		if err != nil {
			io.WriteString(GinkgoWriter, fmt.Sprintf("error reading CNAO logs: %v\n", err))
			continue
		}
		formattedLogs := strings.Replace(string(rawLogs), "\\n", "\n", -1)
		io.WriteString(writer, formattedLogs)
	}
}

func podLogsWriter(namespace string, sinceTime time.Time) func(io.Writer) {
	return func(w io.Writer) {
		writePodsLogs(w, namespace, sinceTime)
	}
}

func writeString(writer io.Writer, message string) {
	writer.Write([]byte(message))
}

func writeMessage(writer io.Writer, message string, args ...string) {
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(formattedMessage, args)
	}
	writeString(writer, formattedMessage)
}
