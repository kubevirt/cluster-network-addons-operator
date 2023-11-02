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
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:9479537fe0827d23bc40056e98f8d1e75778ec294d89ae4d8a62f83dfc74a31d",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:e86b54b96e59679e938ab817d1d69751ce833ee7fa2cc15a09399c74ad715cbf",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:5d9442c26f8750d44f97175f36dbd74bef503f782b9adefcfd08215d065c437a",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:095fc0880882cfdbe1042c7c67e24ef1404720d277d9174a0b231d6ef4e488cd",
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
				Image:      "quay.io/kubevirt/kubemacpool@sha256:095fc0880882cfdbe1042c7c67e24ef1404720d277d9174a0b231d6ef4e488cd",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:b9e211314ed348701ac3ec8d4b95146e66925acae9464bb05fff8644dd6773f4",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:28b255dbb35793226309afdc4dc9465c0ff5d0c3869064a45047c0c312bc18c3",
			},
		},
		SupportedSpec: cnao.NetworkAddonsConfigSpec{
			KubeMacPool: &cnao.KubeMacPool{},
			LinuxBridge: &cnao.LinuxBridge{},
			Multus:      &cnao.Multus{},
			Ovs:         &cnao.Ovs{},
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
