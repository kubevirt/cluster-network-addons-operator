package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
)

func init() {
	release := Release{
		Version: "99.0.0",
		Containers: []cnao.Container{
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "kube-multus",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:3fbcc32bd4e4d15bd93c96def784a229cd84cca27942bf4858b581f31c97ee02",
			},
			{
				ParentName: "dynamic-networks-controller-ds",
				ParentKind: "DaemonSet",
				Name:       "dynamic-networks-controller",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:dee1979d92f0a31598a6e3569ac7004be7d29e7ca9e31db23753ef263110dc04",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:3fbcc32bd4e4d15bd93c96def784a229cd84cca27942bf4858b581f31c97ee02",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:5d24c6d1ecb0556896b7b81c7e5260b54173858425777b7a84df8a706c07e6d2",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:825e3f9fec1996c54a52cec806154945b38f76476b160d554c36e38dfffe5e61",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:ae1288d3450dd79c8d82b2ec0fefa1a8f3a2de17571444484cf28f0d133eb0a6",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "kube-rbac-proxy",
				Image:      components.KubeRbacProxyImageDefault,
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:ae1288d3450dd79c8d82b2ec0fefa1a8f3a2de17571444484cf28f0d133eb0a6",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:b1a73e039f7ef75994e1ae7997a233526dfa808c0260f91de6e1eb9f491fa796",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:b1a73e039f7ef75994e1ae7997a233526dfa808c0260f91de6e1eb9f491fa796",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:e87e829380a1e576384145f78ccaa885ba1d5690d5de7d0b73d40cfb804ea24d",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "secondary-dns",
				Image:      "registry.k8s.io/coredns/coredns@sha256:a0ead06651cf580044aeb0a0feba63591858fb2e43ade8c9dea45a6a89ae7e5e",
			},
		},
		SupportedSpec: cnao.NetworkAddonsConfigSpec{
			KubeMacPool:           &cnao.KubeMacPool{},
			LinuxBridge:           &cnao.LinuxBridge{},
			Multus:                &cnao.Multus{},
			Ovs:                   &cnao.Ovs{},
			MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
			KubeSecondaryDNS:      &cnao.KubeSecondaryDNS{},
		},
		Manifests: []string{
			"network-addons-config.crd.yaml",
			"operator.yaml",
		},
		CrdCleanUp: []string{
			"network-attachment-definitions.k8s.cni.cncf.io",
			"networkaddonsconfigs.networkaddonsoperator.network.kubevirt.io",
		},
	}
	releases = append(releases, release)
}
