package network_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
)

const optInLabelAllowEgressClusterServices = "allow-access-cluster-services"

var netPolTypeMeta = metav1.TypeMeta{
	APIVersion: "networking.k8s.io/v1",
	Kind:       "NetworkPolicy",
}

var _ = Describe("Render Network Policy", func() {
	DescribeTable("should render the network policy according when running on",
		func(info network.ClusterInfo, expected []netv1.NetworkPolicy) {
			obj, err := network.RenderNetworkPolicy("../../data", info)
			Expect(err).NotTo(HaveOccurred())
			bytes, err := json.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())
			var nps []netv1.NetworkPolicy
			Expect(json.Unmarshal(bytes, &nps)).To(Succeed())
			Expect(nps).To(ConsistOf(expected))
		},
		Entry("vanilla k8s",
			network.ClusterInfo{OpenShift4: false},
			testNetworkPolices("kube-system", "k8s-app", "kube-dns", 53),
		),
		Entry("openshift",
			network.ClusterInfo{OpenShift4: true},
			testNetworkPolices("openshift-dns", "dns.operator.openshift.io/daemonset-dns", "default", 5353),
		),
	)
})

func testNetworkPolices(dnsNamespace, dnsLabelKey, dnsLabelValue string, dnsPort int32) []netv1.NetworkPolicy {
	return []netv1.NetworkPolicy{
		newNetPolAllowEgressClusterDNS(dnsNamespace, dnsLabelKey, dnsLabelValue, dnsPort),
		newNetPolAllowEgressClusterAPI(),
	}
}

func newNetPolAllowEgressClusterAPI() netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		TypeMeta:   netPolTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "cnao-allow-egress-to-cluster-api"},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      optInLabelAllowEgressClusterServices,
					Operator: metav1.LabelSelectorOpExists,
				}}},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{{
				Ports: []netv1.NetworkPolicyPort{{
					Protocol: ptr.To(corev1.ProtocolTCP),
					Port:     ptr.To(intstr.FromInt32(6443))}}}},
		},
	}
}
func newNetPolAllowEgressClusterDNS(dnsNamespace, dnsLabelKey, dnsLabelValue string, dnsPort int32) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		TypeMeta:   netPolTypeMeta,
		ObjectMeta: metav1.ObjectMeta{Name: "cnao-allow-egress-to-cluster-dns"},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      optInLabelAllowEgressClusterServices,
					Operator: metav1.LabelSelectorOpExists,
				}}},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress: []netv1.NetworkPolicyEgressRule{{
				To: []netv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{corev1.LabelMetadataName: dnsNamespace},
					},
					PodSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{dnsLabelKey: dnsLabelValue},
					},
				}},
				Ports: []netv1.NetworkPolicyPort{
					{
						Protocol: ptr.To(corev1.ProtocolTCP),
						Port:     ptr.To(intstr.FromInt32(dnsPort)),
					},
					{
						Protocol: ptr.To(corev1.ProtocolUDP),
						Port:     ptr.To(intstr.FromInt32(dnsPort)),
					},
				},
			}},
		},
	}
}
