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
				Image:      "quay.io/kubevirt/cluster-network-addon-multus@sha256:f7350c519d18b2f0fe3ec7acef556b7c42ccdb482ac929a13eb6e66615911da4",
			},
			{
				ParentName: "bridge-marker",
				ParentKind: "DaemonSet",
				Name:       "bridge-marker",
				Image:      "quay.io/kubevirt/bridge-marker@sha256:b6e1537149dbfed3969b23ffc5714b148f8bde86fa5f8c28a5cca7f0f2c28b80",
			},
			{
				ParentName: "kube-cni-linux-bridge-plugin",
				ParentKind: "DaemonSet",
				Name:       "cni-plugins",
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:a90902cf3e5154424148bf3ba3c1bf90316cc77a54042cf6584fe8aedbe6daec",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:c193b76bcd4bbff21ae418b5bef5ca19576fd022ee5fc38c3c0b33c5367dd9ec",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:c193b76bcd4bbff21ae418b5bef5ca19576fd022ee5fc38c3c0b33c5367dd9ec",
			},
			{
				ParentName: "nmstate-handler",
				ParentKind: "DaemonSet",
				Name:       "nmstate-handler",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:7caa956d1fdd289425faec5cfa07a412de3e586c65a5766edf7401c821a979b8",
			},
			{
				ParentName: "nmstate-webhook",
				ParentKind: "Deployment",
				Name:       "nmstate-webhook",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:7caa956d1fdd289425faec5cfa07a412de3e586c65a5766edf7401c821a979b8",
			},
			{
				ParentName: "nmstate-cert-manager",
				ParentKind: "Deployment",
				Name:       "nmstate-cert-manager",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:7caa956d1fdd289425faec5cfa07a412de3e586c65a5766edf7401c821a979b8",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:595df9d7f11bf60c022b66d1bb3ee943a256814b4fa47c1d3fa35f2f487d7a34",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:f989840eee6269abe349a413bdd2b950f206e70650dfe9af56aa7cca26f851ba",
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
