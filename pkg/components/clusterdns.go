package components

// ClusterDNSPlacement describe the cluster DNS information, such as namespace and common label
type ClusterDNSPlacement struct {
	// Namespace is the Namespace name where cluster DNS endpoints reside.
	Namespace string
	// LabelKey is the cluster DNS pods label key that can be used by a label selector.
	LabelKey string
	// LabelValue is the cluster DNS pods label value that can be used by a label selector.
	LabelValue string
}
