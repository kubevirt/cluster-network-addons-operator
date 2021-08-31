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
				Image:      "quay.io/kubevirt/cluster-network-addon-multus@sha256:32867c73cda4d605651b898dc85fea67d93191c47f27e1ad9e9f2b9041c518de",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:9d90a5bd051d71429b6d9fc34112081fe64c6d3fb02221e18ebe72d428d58092",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:d211830c45bde8608befb04cf2accc91490467f3b8e091b812b76453572c9ab8",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:9c885072d4be4924abe542a008b33492aa81806b950f22634c314d258f9b8789",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:9c885072d4be4924abe542a008b33492aa81806b950f22634c314d258f9b8789",
			},
			{
				ParentName: "nmstate-handler",
				ParentKind: "DaemonSet",
				Name:       "nmstate-handler",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:e7fb4ba99567537de9232e90d98c02c7a6eb96a30cb0e131341f74c693efdc5a",
			},
			{
				ParentName: "nmstate-webhook",
				ParentKind: "Deployment",
				Name:       "nmstate-webhook",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:e7fb4ba99567537de9232e90d98c02c7a6eb96a30cb0e131341f74c693efdc5a",
			},
			{
				ParentName: "nmstate-cert-manager",
				ParentKind: "Deployment",
				Name:       "nmstate-cert-manager",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:e7fb4ba99567537de9232e90d98c02c7a6eb96a30cb0e131341f74c693efdc5a",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:1e100c9584044c93c78020b4e4d037f26bbc8cc5b04e51c881ee5d7db5b117fe",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:abf8d51df5904e7a01743524e75c8abdd41922f75fa0093e9fdd01fdfc22ac72",
			},
		},
		SupportedSpec: cnao.NetworkAddonsConfigSpec{
			KubeMacPool: &cnao.KubeMacPool{},
			LinuxBridge: &cnao.LinuxBridge{},
			Multus:      &cnao.Multus{},
			NMState:     &cnao.NMState{},
			Ovs:         &cnao.Ovs{},
		},
		Manifests: []string{
			"network-addons-config.crd.yaml",
			"operator.yaml",
		},
		CrdCleanUp: []string{
			"network-attachment-definitions.k8s.cni.cncf.io",
			"networkaddonsconfigs.networkaddonsoperator.network.kubevirt.io",
			"nodenetworkconfigurationenactments.nmstate.io",
			"nodenetworkconfigurationpolicies.nmstate.io",
			"nodenetworkstates.nmstate.io",
		},
	}
	releases = append(releases, release)
}
