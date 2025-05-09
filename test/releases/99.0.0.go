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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:42ccc54689ea3003d3b6c7decadd85b4e296c15d3ad736da48d7e0c768d1f538",
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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:42ccc54689ea3003d3b6c7decadd85b4e296c15d3ad736da48d7e0c768d1f538",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:ca7fcc798603d9d4de02da00dbda6b0637e72bf1670d531aff070138fded01b4",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:2c602a112411b639f01535f05df4dd977969065f7dbd71821694e81fa617a537",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:2c602a112411b639f01535f05df4dd977969065f7dbd71821694e81fa617a537",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:64baa43b7149add55d7dc814ea180a3bf5480ac44c838cf4b5c4e3fff95aa84c",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:64baa43b7149add55d7dc814ea180a3bf5480ac44c838cf4b5c4e3fff95aa84c",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:13186a0512b59c71e975b4c30e69a6ed0122f83d64da762c7fc5b4a7f066a873",
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
				Image:      "ghcr.io/kubevirt/ipam-controller@sha256:7f9574df14b5b6b0b8e8860a63c658fb683b3f13fdc3223bce5a6cf631c1a142",
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
