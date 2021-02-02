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
				Image:      "quay.io/kubevirt/bridge-marker@sha256:799459f8509a06ea643aa2f6ac02826721a641333aacab07e53166139a5a66c3",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:606eaa06262ad6edd5aa809491c7e46b9ddbc9c52ba8c47e89e734e3e80cd311",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:3efeba40b46b41c8361bf0f247cf98471e70e0dae36fc43eb15d5dbd2e61946e",
			},
			{
				ParentName: "nmstate-handler",
				ParentKind: "DaemonSet",
				Name:       "nmstate-handler",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:3e9748a805d723fef24bc08e0830acec9718f2c7920d7aef6b6eee2e01630e48",
			},
			{
				ParentName: "nmstate-webhook",
				ParentKind: "Deployment",
				Name:       "nmstate-webhook",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:3e9748a805d723fef24bc08e0830acec9718f2c7920d7aef6b6eee2e01630e48",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:d43d34ed4b1bd0b107c2049d21e33f9f870c36e5bf6dc1d80ab567271735c8da",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:8ec72d6a3d58cea27a2b9a52e6ef4ee0d753b3ba989d24555bdfbf01a995a9c1",
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
