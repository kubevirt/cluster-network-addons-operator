package recordingrules

import "github.com/rhobs/operator-observability-toolkit/pkg/operatorrules"

func Register(namespace string) error {
	return operatorrules.RegisterRecordingRules(
		kubemacpoolRecordingRules(namespace),
		operatorRecordingRules(namespace),
	)
}
