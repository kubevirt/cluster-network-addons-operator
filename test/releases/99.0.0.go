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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:fbe669471066f99f37f33bfc362f30f6c473c3850cb21f44813b84a5c3837e8f",
			},
			{
				ParentName: "dynamic-networks-controller-ds",
				ParentKind: "DaemonSet",
				Name:       "dynamic-networks-controller",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:83b460502671fb4f34116363a1a39b2ddfc9d14a920ee0a6413bfc3bd0580404",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:fbe669471066f99f37f33bfc362f30f6c473c3850cb21f44813b84a5c3837e8f",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:18d954d58b9830738df9bf5c9a575d22b33096d1af26fb6bc2da09fb31c9f73a",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:0c354fa9d695b8cab97b459e8afea2f7662407a987e83f6f6f1a8af4b45726be",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:677971a25ff3ce95d9e6ecc86090a09f0ae691ce67b8349384f1881562feed7b",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "kube-rbac-proxy",
				Image:      "quay.io/openshift/origin-kube-rbac-proxy@sha256:e2def4213ec0657e72eb790ae8a115511d5b8f164a62d3568d2f1bff189917e8",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:677971a25ff3ce95d9e6ecc86090a09f0ae691ce67b8349384f1881562feed7b",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:54be8fcacee50af64deafa9e99f3fe079033630c00c4ed9f74d17b0d91009f10",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:54be8fcacee50af64deafa9e99f3fe079033630c00c4ed9f74d17b0d91009f10",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:44489bdf8affb0e82b68a7100bbee7d289151ada54ab2e76a9d3afef769a1f54",
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
				Image:      "ghcr.io/kubevirt/ipam-controller@sha256:aad40edd34f65cf0e087969853d47065aaf411dccf618d196152b583d40300ba",
			},
			{
				ParentName: "passt-binding-cni",
				ParentKind: "DaemonSet",
				Name:       "installer",
				Image:      "ghcr.io/kubevirt/passt-binding-cni@sha256:981c01e0b94ae691ba8ced43c486930085186e9c40b22525c8f0229d1556ee69",
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
