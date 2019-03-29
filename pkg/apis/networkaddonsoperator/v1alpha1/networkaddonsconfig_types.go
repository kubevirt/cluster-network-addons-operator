package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkAddonsConfigSpec defines the desired state of NetworkAddonsConfig
// +k8s:openapi-gen=true
type NetworkAddonsConfigSpec struct {
	Multus          *Multus      `json:"multus,omitempty"`
	LinuxBridge     *LinuxBridge `json:"linuxBridge,omitempty"`
	Sriov           *Sriov       `json:"sriov,omitempty"`
	ImagePullPolicy string       `json:"imagePullPolicy,omitempty"`
}

// +k8s:openapi-gen=true
type Multus struct{}

// +k8s:openapi-gen=true
type LinuxBridge struct{}

// +k8s:openapi-gen=true
type Sriov struct{}

// NetworkAddonsConfigStatus defines the observed state of NetworkAddonsConfig
// +k8s:openapi-gen=true
type NetworkAddonsConfigStatus struct {
	Phase      NetworkAddonsPhase       `json:"phase,omitempty"`
	Conditions []NetworkAddonsCondition `json:"conditions,omitempty" optional:"true"`
}

// NetworkAddonsPhase is a label for the phase of a NetworkAddons deployment at the current time.
// ---
// +k8s:openapi-gen=true
type NetworkAddonsPhase string

// These are the valid NetworkAddons deployment phases
const (
	// The deployment is processing
	NetworkAddonsPhaseDeploying NetworkAddonsPhase = "Deploying"
	// The deployment succeeded
	NetworkAddonsPhaseDeployed NetworkAddonsPhase = "Deployed"
)

// NetworkAddonsCondition represents a condition of a NetworkAddons deployment
// ---
// +k8s:openapi-gen=true
type NetworkAddonsCondition struct {
	Type               NetworkAddonsConditionType `json:"type"`
	Status             corev1.ConditionStatus     `json:"status"`
	LastProbeTime      metav1.Time                `json:"lastProbeTime,omitempty"`
	LastTransitionTime metav1.Time                `json:"lastTransitionTime,omitempty"`
	Reason             string                     `json:"reason,omitempty"`
	Message            string                     `json:"message,omitempty"`
}

// ---
// +k8s:openapi-gen=true
type NetworkAddonsConditionType string

// These are the valid NetworkAddons condition types
const (
	// Whether the deployment or deletion was successful (only used if false)
	NetworkAddonsConditionSynchronized NetworkAddonsConditionType = "Synchronized"
	// Whether all resources were created
	NetworkAddonsConditionCreated NetworkAddonsConditionType = "Created"
	// Whether all components were ready
	NetworkAddonsConditionReady NetworkAddonsConditionType = "Ready"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkAddonsConfig is the Schema for the networkaddonsconfigs API
// +k8s:openapi-gen=true
type NetworkAddonsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkAddonsConfigSpec   `json:"spec,omitempty"`
	Status NetworkAddonsConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkAddonsConfigList contains a list of NetworkAddonsConfig
type NetworkAddonsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkAddonsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkAddonsConfig{}, &NetworkAddonsConfigList{})
}
