package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocpv1 "github.com/openshift/api/config/v1"
	appsv1 "k8s.io/api/apps/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"

	"github.com/kubevirt/cluster-network-addons-operator/test/check"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	"github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

var _ = Describe("TLS", func() {
	BeforeEach(func() {
		By("Create NetworkAddonsConfig deploying all components")
		c := cnao.NetworkAddonsConfigSpec{
			KubeMacPool:            &cnao.KubeMacPool{},
			KubevirtIpamController: &cnao.KubevirtIpamController{},
			LinuxBridge:            &cnao.LinuxBridge{},
			Ovs:                    &cnao.Ovs{},
			MacvtapCni:             &cnao.MacvtapCni{},
			Multus:                 &cnao.Multus{},
			MultusDynamicNetworks:  &cnao.MultusDynamicNetworks{},
			SelfSignConfiguration: &cnao.SelfSignConfiguration{
				// extend certificate intervals for 1 year to relax expired certificate warnings.
				CertRotateInterval:  "8760h",
				CertOverlapInterval: "8760h",
				CAOverlapInterval:   "8760h",
				CARotateInterval:    "8760h",
			},
		}
		operations.CreateConfig(operations.GetCnaoV1GroupVersionKind(), c)
		check.CheckConfigCondition(
			operations.GetCnaoV1GroupVersionKind(),
			check.ConditionAvailable,
			check.ConditionTrue,
			5*time.Minute,
			check.CheckDoNotRepeat,
		)
	})

	AfterEach(func() {
		By("Cleanup")
		gvk := operations.GetCnaoV1GroupVersionKind()
		if operations.GetConfig(gvk) != nil {
			operations.DeleteConfig(gvk)
		}
		check.CheckComponentsRemoval(check.AllComponents)
	})

	It("all services in CNAO namespace should be TLS compliant per TLSComplianceReport", func() {
		const cnaoNamespace = "cluster-network-addons"
		cnaoHost := "cluster-network-addons-operator-prometheus-metrics." + cnaoNamespace
		ipamExtHost := "kubevirt-ipam-controller-webhook-service." + cnaoNamespace
		kmpMetricsHost := "kubemacpool-metrics-service." + cnaoNamespace
		kmpSvcHost := "kubemacpool-service." + cnaoNamespace
		expectedStatus := tlsReportStatus{
			QuantumReady:     true,
			NegotiatedCurves: map[string]string{"TLS 1.3": "X25519MLKEM768"},
			Conditions: []metav1.Condition{
				{Type: "Available", Status: metav1.ConditionTrue},
				{Type: "TLSCompliant", Status: metav1.ConditionTrue},
				{Type: "CertificateValid", Status: metav1.ConditionTrue},
				{Type: "PQCCompliant", Status: metav1.ConditionTrue},
			},
		}
		expectedReports := []tlsReport{
			{
				Spec:   tlsReportSpec{Host: cnaoHost, Port: 8443, SourceNamespace: cnaoNamespace},
				Status: expectedStatus,
			},
			{
				Spec:   tlsReportSpec{Host: ipamExtHost, Port: 443, SourceNamespace: cnaoNamespace},
				Status: expectedStatus,
			},
			{
				Spec:   tlsReportSpec{Host: kmpMetricsHost, Port: 8443, SourceNamespace: cnaoNamespace},
				Status: expectedStatus,
			},
			{
				Spec:   tlsReportSpec{Host: kmpSvcHost, Port: 443, SourceNamespace: cnaoNamespace},
				Status: expectedStatus,
			},
		}
		for reportIndex := range expectedReports {
			expectedReports[reportIndex].Status = cloneTLSReportStatus(expectedStatus)
		}
		expectedHostReports := []string{cnaoHost, ipamExtHost, kmpMetricsHost, kmpSvcHost}

		By("asserting TLS reports")
		var baselineReports []tlsReport
		Eventually(func(g Gomega) {
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())
			actualReports := filterTLSReportsBySpecHost(tlsReports, expectedHostReports...)
			for _, report := range actualReports {
				g.Expect(report.Status.LastError).To(BeEmpty(), "report %s: %+v", report.Spec.Host, report.Status)
			}
			g.Expect(actualReports).To(WithTransform(normalizeConditions, ConsistOf(expectedReports)), "TLS observations (lastCheckAt/checkCount): %+v", actualReports)
			baselineReports = actualReports
		}, 5*time.Minute, 1*time.Second).Should(Succeed())

		By("stopping the scanner to drain in-flight TLS measurements")
		restoreScanner, err := pauseTLSScanner(kubectl.Kubectl)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(restoreScanner)

		By("set tlsSecurityProfile with custom type and explicit non PQC TLS group")
		c := operations.ConvertToConfigV1(operations.GetConfig(operations.GetCnaoV1GroupVersionKind()))
		c.Spec.TLSSecurityProfile = &ocpv1.TLSSecurityProfile{}
		c.Spec.TLSSecurityProfile.Type = ocpv1.TLSProfileCustomType
		c.Spec.TLSSecurityProfile.Custom = &ocpv1.CustomTLSProfile{TLSProfileSpec: ocpv1.TLSProfileSpec{
			MinTLSVersion: ocpv1.VersionTLS13,
			Groups:        []ocpv1.TLSGroup{ocpv1.TLSGroupSecP521r1},
		}}
		operations.UpdateConfig(operations.GetCnaoV1GroupVersionKind(), c.Spec)
		By("waiting for the rendered TLS arguments and complete component rollouts")
		Eventually(func(g Gomega) {
			for _, name := range []string{"kubemacpool-mac-controller-manager", "kubevirt-ipam-controller-manager"} {
				output, stderr, err := kubectl.Kubectl("get", "deployment", name, "-n", cnaoNamespace, "-o", "json", "--request-timeout=15s")
				g.Expect(err).NotTo(HaveOccurred(), stderr)
				var deployment appsv1.Deployment
				g.Expect(json.Unmarshal([]byte(output), &deployment)).To(Succeed())
				g.Expect(tlsProfileRolledOut(deployment, "--tls-group-preferences=CurveP521")).To(Succeed())
			}
		}, 5*time.Minute, time.Second).Should(Succeed())
		By("discarding old observations and starting scans after rollout")
		Expect(deleteTLSReports(kubectl.Kubectl, baselineReports)).To(Succeed())
		Expect(restoreScanner()).To(Succeed())

		By("asserting the specified TLS group is set")
		for i := range expectedReports {
			// expect non PQC compliant state because we configured non PQC TLS group
			expectedReports[i].Status.NegotiatedCurves = map[string]string{"TLS 1.3": "CurveP521"}
			expectedReports[i].Status.QuantumReady = false
			for conditionIndex := range expectedReports[i].Status.Conditions {
				if expectedReports[i].Status.Conditions[conditionIndex].Type == "PQCCompliant" {
					expectedReports[i].Status.Conditions[conditionIndex].Status = metav1.ConditionFalse
				}
			}
		}
		Eventually(func(g Gomega) {
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())
			actualReports := filterTLSReportsBySpecHost(tlsReports, expectedHostReports...)
			for _, report := range actualReports {
				g.Expect(freshTLSReport(report, baselineReports)).To(Succeed())
			}
			g.Expect(actualReports).To(WithTransform(normalizeConditions, ConsistOf(expectedReports)), "TLS observations (lastCheckAt/checkCount): %+v", actualReports)
		}, 5*time.Minute, 1*time.Second).Should(Succeed())
	})
})

func getNamespaceTLSReports(targetNamespace string) ([]tlsReport, error) {
	o, stderr, err := kubectl.Kubectl("get", "tlscompliancereports", "-o", "json", "--request-timeout=15s")
	if err != nil {
		return nil, fmt.Errorf("list TLS reports: %w: %s", err, stderr)
	}

	var tlsReportList tlsReportList
	if err := json.Unmarshal([]byte(o), &tlsReportList); err != nil {
		return nil, err
	}

	var reports []tlsReport
	for _, report := range tlsReportList.Items {
		if report.Spec.SourceNamespace != targetNamespace {
			continue
		}

		reports = append(reports, report)
	}
	return reports, nil
}

func filterTLSReportsBySpecHost(reports []tlsReport, hosts ...string) []tlsReport {
	var filteredReports []tlsReport
	for _, report := range reports {
		if slices.Contains(hosts, report.Spec.Host) {
			filteredReports = append(filteredReports, report)
		}
	}
	return filteredReports
}

func tlsProfileRolledOut(deployment appsv1.Deployment, argument string) error {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if !slices.Contains(container.Args, argument) {
			continue
		}
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 0 &&
			deployment.Status.ObservedGeneration >= deployment.Generation &&
			deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas &&
			deployment.Status.Replicas == *deployment.Spec.Replicas &&
			deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
			return nil
		}
	}
	return fmt.Errorf("deployment %s has not rolled out %s: generation=%d status=%+v", deployment.Name, argument, deployment.Generation, deployment.Status)
}

type tlsKubectl func(...string) (string, string, error)

func runTLSCommand(run tlsKubectl, arguments ...string) error {
	_, stderr, err := run(append(arguments, "--request-timeout=15s")...)
	if err != nil {
		return fmt.Errorf("kubectl %v: %w: %s", arguments, err, stderr)
	}
	return nil
}

func scaleTLSScanner(run tlsKubectl, replicas int32) error {
	if err := runTLSCommand(run, "scale", "deployment/tls-compliance-operator-controller-manager", "-n", "tls-compliance-operator-system", "--replicas="+strconv.FormatInt(int64(replicas), 10)); err != nil {
		return err
	}
	return runTLSCommand(run, "rollout", "status", "deployment/tls-compliance-operator-controller-manager", "-n", "tls-compliance-operator-system", "--timeout=120s")
}

func pauseTLSScanner(run tlsKubectl) (func() error, error) {
	output, stderr, err := run("get", "deployment/tls-compliance-operator-controller-manager", "-n", "tls-compliance-operator-system", "-o", "json", "--request-timeout=15s")
	if err != nil {
		return nil, fmt.Errorf("get scanner deployment: %w: %s", err, stderr)
	}
	var deployment appsv1.Deployment
	if err := json.Unmarshal([]byte(output), &deployment); err != nil {
		return nil, fmt.Errorf("decode scanner deployment: %w", err)
	}
	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas < 1 {
		return nil, fmt.Errorf("expected a running TLS scanner, got replicas=%v", deployment.Spec.Replicas)
	}
	replicas := *deployment.Spec.Replicas
	restore := func() error { return scaleTLSScanner(run, replicas) }
	if err := scaleTLSScanner(run, 0); err != nil {
		return nil, errors.Join(err, restore())
	}
	if err := runTLSCommand(run, "wait", "--for=delete", "pods", "-l", "app.kubernetes.io/name=tls-compliance-operator,control-plane=controller-manager", "-n", "tls-compliance-operator-system", "--timeout=120s"); err != nil {
		return nil, errors.Join(err, restore())
	}
	return restore, nil
}

func deleteTLSReports(run tlsKubectl, reports []tlsReport) error {
	arguments := []string{"delete", "tlscompliancereports"}
	for _, report := range reports {
		if report.Metadata.Name == "" || report.Spec.SourceNamespace != "cluster-network-addons" {
			return fmt.Errorf("refusing to delete report outside the TLS test scope: %+v", report)
		}
		arguments = append(arguments, report.Metadata.Name)
	}
	if len(reports) == 0 {
		return fmt.Errorf("no TLS test reports to delete")
	}
	return runTLSCommand(run, append(arguments, "--ignore-not-found=true", "--timeout=30s")...)
}

func freshTLSReport(report tlsReport, baseline []tlsReport) error {
	var previousCount int64
	var sameReport bool
	for _, previous := range baseline {
		if previous.Spec == report.Spec && previous.Metadata.UID == report.Metadata.UID {
			previousCount = previous.Status.CheckCount
			sameReport = true
			break
		}
	}
	if sameReport || report.Metadata.UID == "" || report.Status.LastCheckAt == nil ||
		report.Status.CheckCount < 1 || report.Status.LastError != "" {
		return fmt.Errorf("report %s:%d uid=%s baselineCheckCount=%d lastCheckAt=%v checkCount=%d lastError=%q",
			report.Spec.Host, report.Spec.Port, report.Metadata.UID, previousCount,
			report.Status.LastCheckAt, report.Status.CheckCount, report.Status.LastError)
	}
	return nil
}

func cloneTLSReportStatus(status tlsReportStatus) tlsReportStatus {
	status.Conditions = slices.Clone(status.Conditions)
	status.NegotiatedCurves = maps.Clone(status.NegotiatedCurves)
	return status
}

// normalizeConditions strip status.Condition non-deterministic fields values.
// Removes all condition's LastTransitionTime, Message and Reason attributes.
func normalizeConditions(reports []tlsReport) []tlsReport {
	clonedReports := slices.Clone(reports)
	// deep copy Conditions because slices.Clone perform shallow copy
	for i := range clonedReports {
		clonedReports[i].Metadata = metav1.ObjectMeta{}
		clonedReports[i].Status.LastCheckAt = nil
		clonedReports[i].Status.CheckCount = 0
		clonedReports[i].Status.LastError = ""
		clonedReports[i].Status.Conditions = slices.Clone(reports[i].Status.Conditions)
	}

	for i, report := range clonedReports {
		for j := range report.Status.Conditions {
			clonedReports[i].Status.Conditions[j].LastTransitionTime = metav1.Time{}
			clonedReports[i].Status.Conditions[j].Message = ""
			clonedReports[i].Status.Conditions[j].Reason = ""
		}
	}

	return clonedReports
}

type tlsReportList struct {
	Items []tlsReport `json:"items"`
}

type tlsReport struct {
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     tlsReportSpec     `json:"spec"`
	Status   tlsReportStatus   `json:"status"`
}
type tlsReportSpec struct {
	SourceNamespace string `json:"sourceNamespace"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
}

type tlsReportStatus struct {
	LastCheckAt      *metav1.Time       `json:"lastCheckAt"`
	CheckCount       int64              `json:"checkCount"`
	LastError        string             `json:"lastError"`
	Conditions       []metav1.Condition `json:"conditions"`
	QuantumReady     bool               `json:"quantumReady"`
	NegotiatedCurves map[string]string  `json:"negotiatedCurves"`
}
