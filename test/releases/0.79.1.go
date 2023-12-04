package releases

import (
	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
)

func init() {
	release := Release{
		Version: "0.79.1",
		Containers: []cnao.Container{
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "kube-multus",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:829c27e9392d013eee5086ca7670d7326d723ebaec526237215e86086b5a3234",
			},
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "install-multus-binary",
				Image:      "ghcr.io/k8snetworkplumbingwg/multus-cni@sha256:829c27e9392d013eee5086ca7670d7326d723ebaec526237215e86086b5a3234",
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
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:5d9442c26f8750d44f97175f36dbd74bef503f782b9adefcfd08215d065c437a",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:9dee66dba5655e2dc70561d4d882d7ee204735ad82f81b24d0b338c428cb918b",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "kube-rbac-proxy",
				Image:      "quay.io/openshift/origin-kube-rbac-proxy@sha256:baedb268ac66456018fb30af395bb3d69af5fff3252ff5d549f0231b1ebb6901",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:9dee66dba5655e2dc70561d4d882d7ee204735ad82f81b24d0b338c428cb918b",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:e95d8f40bf611f1eb3514d50df9195caf9cebfa3716eca26ceafb2e49fefa6da",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:e95d8f40bf611f1eb3514d50df9195caf9cebfa3716eca26ceafb2e49fefa6da",
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
