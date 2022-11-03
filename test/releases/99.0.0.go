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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:4e336bd177b5c60e753be48484abb48edb002c7207de9f265fff2e00e8f5106e",
			},
			{
				ParentName: "dynamic-networks-controller-ds",
				ParentKind: "DaemonSet",
				Name:       "dynamic-networks-controller",
				Image:      "ghcr.io/maiqueb/multus-dynamic-networks-controller@sha256:bd1b07503fd505c66a6ba8b55445a9de94eb322c95d5c22a475df03a8ec67e50",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:4e336bd177b5c60e753be48484abb48edb002c7207de9f265fff2e00e8f5106e",
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
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:c0d14ab010f44bf733aff02b77eb4b5a0ce38fd0c4918a7ecf6941a7bebd72df",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:fb07b1be9e0990e3846ef628e993694bf0765602af5907abf98f7e218db0cb4a",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:fb07b1be9e0990e3846ef628e993694bf0765602af5907abf98f7e218db0cb4a",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:3654b80dd5e459c3e73dd027d732620ed8b488b8a15dfe7922457d16c7e834c3",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:3654b80dd5e459c3e73dd027d732620ed8b488b8a15dfe7922457d16c7e834c3",
			},
		},
		SupportedSpec: cnao.NetworkAddonsConfigSpec{
			KubeMacPool:           &cnao.KubeMacPool{},
			LinuxBridge:           &cnao.LinuxBridge{},
			Multus:                &cnao.Multus{},
			Ovs:                   &cnao.Ovs{},
			MultusDynamicNetworks: &cnao.MultusDynamicNetworks{},
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
