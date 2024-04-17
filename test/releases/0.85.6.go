package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/components"
)

func init() {
	release := Release{
		Version: "0.85.6",
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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:57573a24923e5588bca6bc337a8b2b08406c5b77583974365d2cf063c0dd5d06",
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
				Image:      "quay.io/kubevirt/bridge-marker@sha256:bba066e3b5ff3fb8c5e20861fe8abe51e3c9b50ad6ce3b2616af9cb5479a06d0",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:406b43253fb5d45f50d1543879353822e3f746e2794b65ab30754e800386b76d",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:b84fe57c17cdc07a1c4b5fe26ad63b0fa70bac1e48e218547636992f17935ecb",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:b84fe57c17cdc07a1c4b5fe26ad63b0fa70bac1e48e218547636992f17935ecb",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:08f72edf2bef876bba0b0f5513d30225304ad5e7ad6912a61c083664acdb99ff",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:08f72edf2bef876bba0b0f5513d30225304ad5e7ad6912a61c083664acdb99ff",
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
