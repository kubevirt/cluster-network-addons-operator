package test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	. "github.com/kubevirt/cluster-network-addons-operator/test/check"
	testenv "github.com/kubevirt/cluster-network-addons-operator/test/env"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	. "github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("NMState", func() {
	nmstateVersion := "v0.76.0"
	gvk := GetCnaoV1GroupVersionKind()
	Context("installed as standalone", func() {
		BeforeEach(func() {
			installStandaloneNMState(nmstateVersion)
			checkStandaloneNMStateIsReady(5 * time.Minute)
		})
		Context("and CNAO applies networkaddonsconfig", func() {
			BeforeEach(func() {
				CreateConfig(gvk, cnao.NetworkAddonsConfigSpec{})
				CheckConfigCondition(gvk, ConditionAvailable, ConditionTrue, 15*time.Minute, CheckDoNotRepeat)
			})
			AfterEach(func() {
				DeleteConfig(gvk)
			})

			It("should not remove standalone NMState", func() {
				Eventually(func() error {
					nmstateHandlerDaemonSet := &v1.DaemonSet{}
					return testenv.Client.Get(context.TODO(), types.NamespacedName{Name: "nmstate-handler", Namespace: "nmstate"}, nmstateHandlerDaemonSet)
				}, 5*time.Minute, time.Second).Should(BeNil(), "Timed out waiting for nmstate-operator daemonset")
			})
		})
		AfterEach(func() {
			uninstallStandaloneNMState(nmstateVersion)
		})
	})
})

func installStandaloneNMState(version string) {
	By("Installing standalone kubernetes-nmstate")

	result, stderr, err := kubectl.Kubectl("apply", "-f", fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/crds/nmstate.io_nmstates.yaml", version))
	Expect(err).ToNot(HaveOccurred(), "Error applying CRD: %s", result+stderr)
	result, stderr, err = kubectl.Kubectl("apply", "-f", fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/examples/nmstate.io_v1_nmstate_cr.yaml", version))
	Expect(err).ToNot(HaveOccurred(), "Error applying CR: %s", result+stderr)

	// Create temp directory
	tmpdir, err := ioutil.TempDir("", "operator-test")
	Expect(err).ToNot(HaveOccurred(), "Error creating temporary dir")
	manifests := []string{"namespace", "service_account", "role", "role_binding", "operator"}
	for _, manifest := range manifests {
		yamlString, err := parseManifest(fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/operator/%s.yaml", version, manifest), "latest")
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error parsing manifest file to string: %s", manifest))

		yamlFile := filepath.Join(tmpdir, fmt.Sprintf("%s.yaml", manifest))
		ioutil.WriteFile(yamlFile, []byte(yamlString), 0666)
		result, stderr, err = kubectl.Kubectl("apply", "-f", yamlFile)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error when running kubectl: %s", result+stderr))
	}
}

func uninstallStandaloneNMState(version string) {
	By("Uninstalling kubernetes-nmstate-operator")

	// Create temp directory
	tmpdir, err := ioutil.TempDir("", "operator-test")
	Expect(err).ToNot(HaveOccurred(), "Error creating temporary dir")
	manifests := []string{"operator", "role_binding", "role", "service_account", "namespace"}
	for _, manifest := range manifests {
		yamlString, err := parseManifest(fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/operator/%s.yaml", version, manifest), version)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error parsing manifest file to string: %s", manifest))

		yamlFile := filepath.Join(tmpdir, fmt.Sprintf("%s.yaml", manifest))
		ioutil.WriteFile(yamlFile, []byte(yamlString), 0666)
		result, stderr, err := kubectl.Kubectl("delete", "-f", yamlFile)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Error when running kubectl: %s", result+stderr))
	}
	result, stderr, err := kubectl.Kubectl("delete", "-f", fmt.Sprintf("https://raw.githubusercontent.com/nmstate/kubernetes-nmstate/%s/deploy/crds/nmstate.io_nmstates.yaml", version))
	Expect(err).ToNot(HaveOccurred(), "Error deleting CRD: %s", result+stderr)
}

func parseManifest(url string, tag string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "Could not get url: %s", url)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "Error reading body of url: %s", url)
	}
	var tmpl *template.Template
	tmpl = template.Must(template.New("manifest").Parse(string(body)))

	data := struct {
		OperatorNamespace  string
		OperatorImage      string
		OperatorPullPolicy string
		HandlerNamespace   string
		HandlerImage       string
		HandlerPullPolicy  string
	}{
		OperatorNamespace:  "nmstate",
		OperatorImage:      fmt.Sprintf("quay.io/nmstate/kubernetes-nmstate-operator:%s", tag),
		OperatorPullPolicy: "Always",
		HandlerNamespace:   "nmstate",
		HandlerImage:       fmt.Sprintf("quay.io/nmstate/kubernetes-nmstate-handler:%s", tag),
		HandlerPullPolicy:  "Always",
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", errors.Wrapf(err, "Error parsing template")
	}
	return out.String(), nil
}

func checkStandaloneNMStateIsReady(timeout time.Duration) {
	By("Checking that the operator is up and running")
	if timeout != CheckImmediately {
		Eventually(func() error {
			return CheckForGenericDeployment("nmstate-webhook", "nmstate", false, false)
		}, timeout, time.Second).ShouldNot(HaveOccurred(), fmt.Sprintf("Timed out waiting for the nmstate-webhook to become ready"))
	} else {
		Expect(CheckForGenericDeployment("nmstate-webook", "nmstate", false, false)).ShouldNot(HaveOccurred(), "nmstate-webhook is not ready")
	}
}
