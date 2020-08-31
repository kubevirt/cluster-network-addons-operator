package releases

import (
	opv1alpha1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1alpha1"
)

func init() {
	release := Release{
		Version: "99.0.0",
		Containers: []opv1alpha1.Container{
			{
				ParentName: "multus",
				ParentKind: "DaemonSet",
				Name:       "kube-multus",
				Image:      "nfvpe/multus@sha256:167722b954355361bd69829466f27172b871dbdbf86b85a95816362885dc0aba",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:e55f73526468fee46a35ae41aa860f492d208b8a7a132832c5b9a76d4a51566a",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:680ac8fd5eeab39c9a3c01479da344bdcaa43aa065d07ae00513b7bafa22fccf",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:ad8ca6d379d495804969ba4d03da9a6936ff8f413f6f6c7bd20e0138dc0303c4",
			},
			{
				ParentName: "nmstate-handler",
				ParentKind: "DaemonSet",
				Name:       "nmstate-handler",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30",
			},
			{
				ParentName: "nmstate-webhook",
				ParentKind: "Deployment",
				Name:       "nmstate-webhook",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:76cc13fb4a60943dca6038619599b6a49fe451852aba23ad3046658429a9af30",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:4101c52617efb54a45181548c257a08e3689f634b79b9dfcff42bffd8b25af53",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:0f08d6b1550a90c9f10221f2bb07709d1090e7c675ee1a711981bd429074d620",
			},
		},
		SupportedSpec: opv1alpha1.NetworkAddonsConfigSpec{
			KubeMacPool: &opv1alpha1.KubeMacPool{},
			LinuxBridge: &opv1alpha1.LinuxBridge{},
			Multus:      &opv1alpha1.Multus{},
			NMState:     &opv1alpha1.NMState{},
			Ovs:         &opv1alpha1.Ovs{},
		},
		Manifests: []string{
			"operator.yaml",
		},
	}
	releases = append(releases, release)
}
