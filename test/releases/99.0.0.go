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
				Image:      "quay.io/kubevirt/cni-default-plugins@sha256:5d9442c26f8750d44f97175f36dbd74bef503f782b9adefcfd08215d065c437a",
			},
			{
				ParentName: "kubemacpool-mac-controller-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:5819cc8b64fde5d103a24a0426545bce9ff53645e1dd727fc941199905fdfd26",
			},
			{
				ParentName: "kubemacpool-cert-manager",
				ParentKind: "Deployment",
				Name:       "manager",
				Image:      "quay.io/kubevirt/kubemacpool@sha256:5819cc8b64fde5d103a24a0426545bce9ff53645e1dd727fc941199905fdfd26",
			},
			{
				ParentName: "nmstate-handler",
				ParentKind: "DaemonSet",
				Name:       "nmstate-handler",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:2ec6625078a1013a9d39269980318ccb64c6bcf1cb603793e4044880ffecf921",
			},
			{
				ParentName: "nmstate-webhook",
				ParentKind: "Deployment",
				Name:       "nmstate-webhook",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:2ec6625078a1013a9d39269980318ccb64c6bcf1cb603793e4044880ffecf921",
			},
			{
				ParentName: "nmstate-cert-manager",
				ParentKind: "Deployment",
				Name:       "nmstate-cert-manager",
				Image:      "quay.io/nmstate/kubernetes-nmstate-handler@sha256:2ec6625078a1013a9d39269980318ccb64c6bcf1cb603793e4044880ffecf921",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-plugin",
				Image:      "quay.io/kubevirt/ovs-cni-plugin@sha256:caabc3fb82af233c9d55cf96c38e48ded67bd339f449231c32e92457671e2f45",
			},
			{
				ParentName: "ovs-cni-amd64",
				ParentKind: "DaemonSet",
				Name:       "ovs-cni-marker",
				Image:      "quay.io/kubevirt/ovs-cni-marker@sha256:1d844daae0f7a715a2d2e6a3490d3643cba7fd99d85205de09d629e89e05915f",
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
