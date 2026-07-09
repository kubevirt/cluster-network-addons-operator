package test

import (
	"encoding/json"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
			KubeSecondaryDNS:       &cnao.KubeSecondaryDNS{},
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
		expectedStatus := tlsReportStatus{
			QuantumReady: true,
			Conditions: []metav1.Condition{
				{
					Type:   "Available",
					Status: metav1.ConditionTrue,
					Reason: "EndpointDiscovered",
				},
				{
					Type:    "TLSCompliant",
					Status:  metav1.ConditionTrue,
					Reason:  "Compliant",
					Message: "Endpoint supports modern TLS (1.2 or 1.3)",
				},
				{
					Type:   "CertificateValid",
					Status: "True",
					Reason: "Valid",
				},
			},
		}
		expectedReports := []tlsReport{
			{
				Spec: tlsReportSpec{
					Host:            "cluster-network-addons-operator-prometheus-metrics." + cnaoNamespace,
					Port:            8443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            "kubevirt-ipam-controller-webhook-service." + cnaoNamespace,
					Port:            443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            "kubemacpool-metrics-service." + cnaoNamespace,
					Port:            8443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            "kubemacpool-service." + cnaoNamespace,
					Port:            443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
		}

		Eventually(func(g Gomega) {
			var err error
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(tlsReports).To(
				WithTransform(
					// normalizeConditions is used because conditions contains non-deterministic
					// values (such as LastTransitionTime, Message) breaking underlying equality matchers.
					normalizeConditions,
					// ContainElements is used due to existing reports for arbitrary pods exposing some ports.
					// The test care about endpoints CNAO deploy and advertise via Services.
					ContainElements(expectedReports)))
		}, 5*time.Minute, 1*time.Second).Should(Succeed())
	})
})

func getNamespaceTLSReports(targetNamespace string) ([]tlsReport, error) {
	o, _, err := kubectl.Kubectl("get", "tlscompliancereports", "-o", "json")
	if err != nil {
		return nil, err
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

// normalizeConditions strip status.Condition non-deterministic fields values.
// Removes all conditions LastTransitionTime.
// For condition of type 'Available' and 'CertificateValid' the 'Message' field is trimmed
// because it contains non-deterministic text.
func normalizeConditions(reports []tlsReport) []tlsReport {
	clonedReports := slices.Clone(reports)
	// deep copy Conditions because slices.Clone perform shallow copy
	for i := range clonedReports {
		clonedReports[i].Status.Conditions = slices.Clone(reports[i].Status.Conditions)
	}

	for i, report := range clonedReports {
		for j := range report.Status.Conditions {
			clonedReports[i].Status.Conditions[j].LastTransitionTime = metav1.Time{}
		}
	}

	// The following conditions type message's containing non-deterministic text,
	// hence remove Message field value
	for i, report := range clonedReports {
		for j, condition := range report.Status.Conditions {
			if condition.Type == "Available" || condition.Type == "CertificateValid" {
				clonedReports[i].Status.Conditions[j].Message = ""
			}
		}
	}

	return clonedReports
}

type tlsReportList struct {
	Items []tlsReport `json:"items"`
}

type tlsReport struct {
	Spec   tlsReportSpec   `json:"spec"`
	Status tlsReportStatus `json:"status"`
}
type tlsReportSpec struct {
	SourceNamespace string `json:"sourceNamespace"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
}

type tlsReportStatus struct {
	Conditions   []metav1.Condition `json:"conditions"`
	QuantumReady bool               `json:"quantumReady"`
}
