package test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"

	"github.com/kubevirt/cluster-network-addons-operator/test/check"
	"github.com/kubevirt/cluster-network-addons-operator/test/kubectl"
	"github.com/kubevirt/cluster-network-addons-operator/test/operations"
)

const (
	tlsReportResource         = "tlscompliancereports"
	complianceStatusKey       = "complianceStatus"
	complianceStatusCompliant = "Compliant"

	cnaoNamespace = "cluster-network-addons"
)

var _ = Describe("TLS Compliance", func() {
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
		Eventually(func(g Gomega) {
			var err error
			tlsReports, err := getNamespaceTLSReports(cnaoNamespace)
			g.Expect(err).NotTo(HaveOccurred())

			for _, report := range tlsReports {
				statusRaw := report.Status[complianceStatusKey]
				status := statusRaw.(string)
				g.Expect(status).To(Equal(complianceStatusCompliant))
			}
		}, 5*time.Minute, 1*time.Second).Should(Succeed())
	})
})

type tlsReport struct {
	Spec   map[string]interface{} `json:"spec"`
	Status map[string]interface{} `json:"status"`
}

type tlsReportList struct {
	Items []tlsReport `json:"items"`
}

func getNamespaceTLSReports(targetNamespace string) ([]tlsReport, error) {
	o, _, err := kubectl.Kubectl("get", tlsReportResource, "-o", "json")
	if err != nil {
		return nil, err
	}

	var tlsReportList tlsReportList
	if err := json.Unmarshal([]byte(o), &tlsReportList); err != nil {
		return nil, err
	}
	
	const sourceNamespaceKey = "sourceNamespace"

	var reports []tlsReport
	for _, report := range tlsReportList.Items {
		if report.Spec == nil {
			continue
		}
		srcNamespace, _ := report.Spec[sourceNamespaceKey].(string)
		if srcNamespace != targetNamespace {
			continue
		}

		reports = append(reports, report)
	}
	return reports, nil
}
