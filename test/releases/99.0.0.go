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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:e93af9e533cc16ca7847aeec3b516c9b1e8891c2b95c727991e94c9b59c4970e",
			},
			{
				ParentName: "dynamic-networks-controller-ds",
				ParentKind: "DaemonSet",
				Name:       "dynamic-networks-controller",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:2a2bb32c0ea8b232b3dbe81c0323a107e8b05f8cad06704fca2efd0d993a87be",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:e93af9e533cc16ca7847aeec3b516c9b1e8891c2b95c727991e94c9b59c4970e",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:f9611ec10bb4aec44b0ec19f9b9d748a36255c089a1f59bc76e5fc37acc0fed2",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:dec803245db555f7f0aca4d0fb70c1463a78f046a331efd10d6e0769d4fa1915",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:dec803245db555f7f0aca4d0fb70c1463a78f046a331efd10d6e0769d4fa1915",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:435f374b434b3bc70a5cfaba0011fdcf5f433d96b98b06d29306cbd8db3a8c21",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "ghcr.io/k8snetworkplumbingwg/ovs-cni-plugin@sha256:435f374b434b3bc70a5cfaba0011fdcf5f433d96b98b06d29306cbd8db3a8c21",
			},
			{
				ParentName: "secondary-dns",
				ParentKind: "Deployment",
				Name:       "status-monitor",
				Image:      "ghcr.io/kubevirt/kubesecondarydns@sha256:f5fe9c98fb6d7e5e57a6df23fe82e43e65db5953d76af44adda9ab40c46ad0bf",
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
				Image:      "ghcr.io/kubevirt/ipam-controller@sha256:69b25271d1ad2521f38f5fa8b9623973ad7c74f33f88d17c0ebc9882ef9c2c5e",
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
