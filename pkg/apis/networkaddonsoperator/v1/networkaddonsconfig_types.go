package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetworkAddonsConfigSpec defines the desired state of NetworkAddonsConfig
type NetworkAddonsConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// NetworkAddonsConfigStatus defines the observed state of NetworkAddonsConfig
type NetworkAddonsConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetworkAddonsConfig is the Schema for the networkaddonsconfigs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=networkaddonsconfigs,scope=Namespaced
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
