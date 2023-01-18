package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
)

func init() {
	release := Release{
		Version: "0.85.0",
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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:ef8fe97a52eb9b3c03e99979a42cf2edaa7b3365cb3eb4dd1654b1bb9e73d7a3",
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
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:e75d67fb44f4b748137c85b1c06b500607410a3218cafbf8a4bb6359f2b90373",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:0cc5ad824fc163d6dea5e9bd872467c691eaa9a88944008b5d746495b2a72214",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:0cc5ad824fc163d6dea5e9bd872467c691eaa9a88944008b5d746495b2a72214",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:5f7290e2294255ab2547c3b4bf48cc2d75531ec5a43e600366e9b2719bef983f",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:5f7290e2294255ab2547c3b4bf48cc2d75531ec5a43e600366e9b2719bef983f",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:b489a7c5d05b000f776c9c302985b8b7f29ff31f577a1480912ed625c8772d6b",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "secondary-dns",
				Image:      components.CoreDNSImageDefault,
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
