package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

func init() {
	release := Release{
		Version: "0.91.1",
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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-dynamic-networks-controller@sha256:57573a24923e5588bca6bc337a8b2b08406c5b77583974365d2cf063c0dd5d06",
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
				Image:      "quay.io/kubevirt/bridge-marker@sha256:bba066e3b5ff3fb8c5e20861fe8abe51e3c9b50ad6ce3b2616af9cb5479a06d0",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:c884d6d08f8c0db98964f1eb3877b44ade41fa106083802a9914775df17d5291",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:d9ca6b68dfd9b9da00f7f908c693eb2dc042ef5bc7dab870c768b93ae3170fa4",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:d9ca6b68dfd9b9da00f7f908c693eb2dc042ef5bc7dab870c768b93ae3170fa4",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:e16ac74343da21abb8fb668ce71e728053d00503a992dae2164b9e94a280113e",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:e16ac74343da21abb8fb668ce71e728053d00503a992dae2164b9e94a280113e",
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
