package test

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocpv1 "github.com/openshift/api/config/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"

	"github.com/kubevirt/cluster-network-addons-operator/test/check"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	"github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

const (
	conditionTypeAvailable        = "Available"
	conditionTypeTLSCompliant    = "TLSCompliant"
	conditionTypeCertificateValid = "CertificateValid"
	conditionTypePQCCompliant    = "PQCCompliant"

	conditionReasonEndpointDiscovered = "EndpointDiscovered"
	conditionReasonCompliant          = "Compliant"
	conditionReasonValid              = "Valid"
	conditionReasonPQCReady           = "PQCReady"

	pqcLegacyMessageSubstring = "post-quantum key exchange (ML-KEM)"
	pqcHybridMessageSubstring = "hybrid ML-KEM key exchange"
	pqcActiveProbeSuffix      = " (verified by active probe)"
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
		cnaoHost := "cluster-network-addons-operator-prometheus-metrics." + cnaoNamespace
		ipamExtHost := "kubevirt-ipam-controller-webhook-service." + cnaoNamespace
		kmpMetricsHost := "kubemacpool-metrics-service." + cnaoNamespace
		kmpSvcHost := "kubemacpool-service." + cnaoNamespace

		expectedStatus := tlsReportStatus{
			QuantumReady:     true,
			NegotiatedCurves: map[string]string{"TLS 1.3": "X25519MLKEM768"},
			Conditions: []metav1.Condition{
				{
					Type:   conditionTypeAvailable,
					Status: metav1.ConditionTrue,
					Reason: conditionReasonEndpointDiscovered,
				},
				{
					Type:    conditionTypeTLSCompliant,
					Status:  metav1.ConditionTrue,
					Reason:  conditionReasonCompliant,
					Message: "Endpoint supports modern TLS (1.2 or 1.3)",
				},
				{
					Type:   conditionTypeCertificateValid,
					Status: metav1.ConditionTrue,
					Reason: conditionReasonValid,
				},
				{
					Type:    conditionTypePQCCompliant,
					Status:  metav1.ConditionTrue,
					Reason:  conditionReasonPQCReady,
					Message: "Endpoint supports TLS 1.3 with " + pqcHybridMessageSubstring,
				},
			},
		}
		expectedReports := []tlsReport{
			{
				Spec: tlsReportSpec{
					Host:            cnaoHost,
					Port:            8443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            ipamExtHost,
					Port:            443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            kmpMetricsHost,
					Port:            8443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
			{
				Spec: tlsReportSpec{
					Host:            kmpSvcHost,
					Port:            443,
					SourceNamespace: cnaoNamespace,
				},
				Status: expectedStatus,
			},
		}
		By("asserting TLS reports")
		Eventually(func(g Gomega) {
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())
			actualReports := filterTLSReportsBySpecHost(tlsReports, cnaoHost, ipamExtHost, kmpMetricsHost, kmpSvcHost)
			g.Expect(actualReports).To(WithTransform(normalizeConditions, ConsistOf(expectedReports)))
		}, 5*time.Minute, 1*time.Second).Should(Succeed())

		By("set tlsSecurityProfile with custom type and explicit TLS group")
		c := operations.ConvertToConfigV1(operations.GetConfig(operations.GetCnaoV1GroupVersionKind()))
		c.Spec.TLSSecurityProfile = &ocpv1.TLSSecurityProfile{}
		c.Spec.TLSSecurityProfile.Type = ocpv1.TLSProfileCustomType
		c.Spec.TLSSecurityProfile.Custom = &ocpv1.CustomTLSProfile{TLSProfileSpec: ocpv1.TLSProfileSpec{
			MinTLSVersion: ocpv1.VersionTLS13,
			Groups:        []ocpv1.TLSGroup{ocpv1.TLSGroupSecP521r1},
		}}
		operations.UpdateConfig(operations.GetCnaoV1GroupVersionKind(), c.Spec)
		check.CheckConfigCondition(
			operations.GetCnaoV1GroupVersionKind(),
			check.ConditionAvailable,
			check.ConditionTrue,
			5*time.Minute,
			check.CheckDoNotRepeat,
		)

		// TODO: remove below slice and assert all tested components once they are all wired to tlsSecurityProfile groups
		tlsGroupSupportedHosts := []string{cnaoHost, kmpMetricsHost, kmpSvcHost}

		By("asserting the specified TLS group is set")
		expectedNegotiatedCurves := map[string]string{"TLS 1.3": "CurveP521"}
		Eventually(func(g Gomega) {
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())
			actualReports := filterTLSReportsBySpecHost(tlsReports, tlsGroupSupportedHosts...)
			g.Expect(actualReports).To(HaveLen(len(tlsGroupSupportedHosts)))
			for _, report := range actualReports {
				g.Expect(report.Status.NegotiatedCurves).To(Equal(expectedNegotiatedCurves))
			}
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

func filterTLSReportsBySpecHost(reports []tlsReport, hosts ...string) []tlsReport {
	var filteredReports []tlsReport
	for _, report := range reports {
		if slices.Contains(hosts, report.Spec.Host) {
			filteredReports = append(filteredReports, report)
		}
	}
	return filteredReports
}

// normalizeConditions strip status.Condition non-deterministic fields values.
// Removes all conditions LastTransitionTime.
// For condition of type 'Available' and 'CertificateValid' the 'Message' field is trimmed
// because it contains non-deterministic text.
func normalizeConditions(reports []tlsReport) []tlsReport {
	clonedReports := slices.Clone(reports)
	for i := range clonedReports {
		// deep copy Conditions because slices.Clone perform shallow copy
		clonedReports[i].Status.Conditions = slices.Clone(reports[i].Status.Conditions)
		for j := range clonedReports[i].Status.Conditions {
			condition := &clonedReports[i].Status.Conditions[j]
			condition.LastTransitionTime = metav1.Time{}
			condition.ObservedGeneration = 0

			switch condition.Type {
			case conditionTypeAvailable, conditionTypeCertificateValid:
				condition.Message = ""
			case conditionTypePQCCompliant:
				msg := strings.ReplaceAll(condition.Message, pqcLegacyMessageSubstring, pqcHybridMessageSubstring)
				condition.Message = strings.TrimSuffix(msg, pqcActiveProbeSuffix)
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
	Conditions       []metav1.Condition `json:"conditions"`
	QuantumReady     bool               `json:"quantumReady"`
	NegotiatedCurves map[string]string  `json:"negotiatedCurves"`
}
