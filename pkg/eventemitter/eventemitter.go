package eventemitter

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	"github.com/kubevirt/cluster-network-addons-operator/pkg/names"
)

const (
	ModifiedMessage    = "Config spec was modified"
	ProgressingMessage = "Operator is progressing"
	AvailableMessage   = "All applied components are available"
	FailedMessage      = "Operator failed"
)

const (
	AvailableReason   = "Available"
	ProgressingReason = "Progressing"
	FailedReason      = "Failed"
	ModifiedReason    = "Modified"
)

type EventEmitter interface {
	Init(mgr manager.Manager)
	EmitEventForConfig(config *cnaov1.NetworkAddonsConfig, eventType, reason, msg string)
	EmitModifiedForConfig()
	EmitProgressingForConfig()
	EmitFailingForConfig(reason, message string)
	EmitAvailableForConfig()
}

type eventEmitter struct {
	recorder record.EventRecorder
	client   client.Client
}

// New event emitter
func New(mgr manager.Manager) EventEmitter {
	var evntEmtr EventEmitter = &eventEmitter{
		client: mgr.GetClient(),
	}

	evntEmtr.Init(mgr)
	return evntEmtr
}

func (ee *eventEmitter) Init(mgr manager.Manager) {
	ee.recorder = mgr.GetEventRecorderFor(names.OperatorConfig)
}

func (ee eventEmitter) EmitEventForConfig(config *cnaov1.NetworkAddonsConfig, eventType, reason, msg string) {
	if config != nil {
		ee.recorder.Event(config, eventType, reason, msg)
	}
}

func (ee eventEmitter) EmitProgressingForConfig() {
	config := ee.getConfigForEmitter()
	ee.EmitEventForConfig(config, corev1.EventTypeNormal, ProgressingReason, ProgressingMessage)
}

func (ee eventEmitter) EmitFailingForConfig(reason, message string) {
	config := ee.getConfigForEmitter()
	ee.EmitEventForConfig(config, corev1.EventTypeWarning, fmt.Sprintf("%s: %s", FailedReason, reason), fmt.Sprintf("%s: %s", FailedMessage, message))
}

func (ee eventEmitter) EmitAvailableForConfig() {
	config := ee.getConfigForEmitter()
	ee.EmitEventForConfig(config, corev1.EventTypeNormal, AvailableReason, AvailableMessage)
}

func (ee eventEmitter) EmitModifiedForConfig() {
	config := ee.getConfigForEmitter()
	ee.EmitEventForConfig(config, corev1.EventTypeNormal, ModifiedReason, ModifiedMessage)
}

func (ee eventEmitter) getConfigForEmitter() *cnaov1.NetworkAddonsConfig {
	config := &cnaov1.NetworkAddonsConfig{}
	err := ee.client.Get(context.TODO(), types.NamespacedName{Name: names.OperatorConfig}, config)
	if err != nil {
		log.Printf("Failed to get NetworkAddonsConfig in order emit event. %v", err)

		return nil
	}

	return config
}
