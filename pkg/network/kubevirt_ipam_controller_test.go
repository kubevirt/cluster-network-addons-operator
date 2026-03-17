package network_test

import (
	"fmt"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ocpv1 "github.com/openshift/api/config/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/network"
)

var _ = Describe("RenderKubevirtIPAMController", func() {
	const manifestDir = "../../data"

	var clusterInfo = &network.ClusterInfo{}

	DescribeTable("should set TLS settings flags correctly, given TLS profile is",
		func(testTLSSecurityProfile *ocpv1.TLSSecurityProfile, expectedFlags []string) {
			conf := &cnao.NetworkAddonsConfigSpec{
				PlacementConfiguration: &cnao.PlacementConfiguration{
					Infra:     &cnao.Placement{NodeSelector: map[string]string{"test": "infra"}},
					Workloads: &cnao.Placement{NodeSelector: map[string]string{"test": "workers"}},
				},
				ImagePullPolicy:        v1.PullAlways,
				TLSSecurityProfile:     testTLSSecurityProfile,
				KubevirtIpamController: &cnao.KubevirtIpamController{DefaultNetworkNADNamespace: "ovn-kubernetes"},
			}

			objs, err := network.RenderKubevirtIPAMController(conf, manifestDir, clusterInfo)
			Expect(err).NotTo(HaveOccurred())
			Expect(objs).NotTo(BeEmpty())
			deployment, err := getDeployment(objs...)
			Expect(err).NotTo(HaveOccurred())
			Expect(deployment.Spec.Template.Spec.Containers[0].Args).To(ContainElements(expectedFlags))
		},
		Entry("Old",
			&ocpv1.TLSSecurityProfile{Type: ocpv1.TLSProfileOldType},
			[]string{
				"--tls-min-version=VersionTLS10",
				"--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256," +
					"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256," +
					"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256," +
					"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA," +
					"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_CBC_SHA256," +
					"TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_3DES_EDE_CBC_SHA",
			},
		),
		Entry("Intermediate",
			&ocpv1.TLSSecurityProfile{Type: ocpv1.TLSProfileIntermediateType},
			[]string{
				"--tls-min-version=VersionTLS12",
				"--tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384," +
					"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
		),
		Entry("Modern",
			&ocpv1.TLSSecurityProfile{Type: ocpv1.TLSProfileModernType},
			[]string{"--tls-min-version=VersionTLS13"},
		),
		Entry("Custom, min TLS version 1.3 and ciphers, should not specify ciphers flag",
			&ocpv1.TLSSecurityProfile{
				Type: ocpv1.TLSProfileCustomType,
				Custom: &ocpv1.CustomTLSProfile{
					TLSProfileSpec: ocpv1.TLSProfileSpec{
						MinTLSVersion: ocpv1.VersionTLS13,
						Ciphers:       []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
					},
				},
			},
			[]string{"--tls-min-version=VersionTLS13"},
		),
		Entry("Custom, min TLS version 1.2 and ciphers",
			&ocpv1.TLSSecurityProfile{
				Type: ocpv1.TLSProfileCustomType,
				Custom: &ocpv1.CustomTLSProfile{
					TLSProfileSpec: ocpv1.TLSProfileSpec{
						MinTLSVersion: ocpv1.VersionTLS12,
						Ciphers:       []string{"AES128-SHA", "AES256-SHA"},
					},
				},
			},
			[]string{
				"--tls-min-version=VersionTLS12",
				"--tls-cipher-suites=TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA",
			},
		),
	)
})

func getDeployment(objs ...*unstructured.Unstructured) (*appsv1.Deployment, error) {
	idx := slices.IndexFunc(objs, func(obj *unstructured.Unstructured) bool {
		return obj.GetKind() == "Deployment"
	})
	if idx == -1 {
		return nil, fmt.Errorf("deployment not found")
	}
	deployment := &appsv1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(objs[idx].Object, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}
