package names

// NAME is the default name of the operator, it can be changed by
// manifest-templator
const NAME = "cluster-network-addons-operator"

// NAMESPACE is the default namespace of the operator, it can be changed by the
// manifest templator
const NAMESPACE = "cluster-network-addons"

// OPERATOR_CONFIG is the name of the CRD that defines the complete
// operator configuration
const OPERATOR_CONFIG = "cluster"

// APPLIED_PREFIX is the prefix applied to the config maps
// where we store previously applied configuration
const APPLIED_PREFIX = "cluster-networks-addons-operator-applied-"
