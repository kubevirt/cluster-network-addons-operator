package network

type ClusterInfo struct {
	SCCAvailable           bool
	OpenShift4             bool
	NmstateOperator        bool
	NmstateCRExists        bool
	MonitoringAvailable    bool
	IsSingleReplica        bool
	ClusterReaderAvailable bool
}
