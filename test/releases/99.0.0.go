package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

func init() {
	release := Release{
		Version: "99.0.0",
		Containers: []cnao.Container{
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "kube-multus",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:16030e8320088cf74dc5fbc7dccfea40169f09722dfad15765fa28bceb8de439",
			},
			{
				ParentName: "dynamic-networks-controller-ds",
				ParentKind: "DaemonSet",
				Name:       "dynamic-networks-controller",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:8061bd1276ff022fe52a0b07bc6fa8d27e5f6f20cf3bf764e76d347d2e3c929b",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:16030e8320088cf74dc5fbc7dccfea40169f09722dfad15765fa28bceb8de439",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:e492ca4a6d1234781928aedefb096941d95babee4116baaba4d2a3834813826a",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:976a24392c2a096c38c2663d234b2d3131f5c24558889196d30b9ac1b6716788",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:477a781717cebdb3ce7b63338c44abb9e34058ca9849f932d6d5cef8950b532e",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "kube-rbac-proxy",
				Image:      "quay.io/brancz/kube-rbac-proxy@sha256:e6a323504999b2a4d2a6bf94f8580a050378eba0900fd31335cf9df5787d9a9b",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:477a781717cebdb3ce7b63338c44abb9e34058ca9849f932d6d5cef8950b532e",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:d088e47f181007fe4823f0384ebae071950d105cd36c9187f9d06fd815288990",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:d088e47f181007fe4823f0384ebae071950d105cd36c9187f9d06fd815288990",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:8273cdbc438e06864eaa8e47947bea18fa5118a97cdaddc41b5dfa6e13474c79",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "secondary-dns",
				Image:      "registry.k8s.io/coredns/coredns@sha256:a0ead06651cf580044aeb0a0feba63591858fb2e43ade8c9dea45a6a89ae7e5e",
			},
			{
				ParentName: "kubevirt-ipam-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "ghcr.io/kubevirt/ipam-controller@sha256:b1a25c8f60775e176c3dd0911fe9ffec0f5bf2f5687d2d3aee5c4d1702532071",
			},
			{
				ParentName: "passt-binding-cni",
				ParentKind: "DaemonSet",
				Name:       "installer",
				Image:      "ghcr.io/kubevirt/passt-binding-cni@sha256:7ebfa89a30965c5297f86a12cf799346a71439a197553eca322e064c5d8e1ff0",
			},
		},
		SupportedSpec: cnao.NetworkAddonsConfigSpec{
			KubeMacPool:            &cnao.KubeMacPool{},
			LinuxBridge:            &cnao.LinuxBridge{},
			Multus:                 &cnao.Multus{},
			Ovs:                    &cnao.Ovs{},
			MultusDynamicNetworks:  &cnao.MultusDynamicNetworks{},
			KubeSecondaryDNS:       &cnao.KubeSecondaryDNS{},
			KubevirtIpamController: &cnao.KubevirtIpamController{},
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
