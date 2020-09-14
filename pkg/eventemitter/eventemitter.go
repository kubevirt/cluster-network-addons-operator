package eventemitter

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	EmitEvent(object runtime.Object, eventType, reason, msg string)
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
	ee.recorder = mgr.GetEventRecorderFor(names.OPERATOR_CONFIG)
}

func (ee eventEmitter) EmitEvent(object runtime.Object, eventType, reason, msg string) {
	if object != nil {
		ee.recorder.Event(object, eventType, reason, msg)
	}
}

func (ee eventEmitter) EmitProgressingForConfig() {
	config, err := ee.getConfigForEmitter()
	if err != nil {
		log.Printf("Failed to emit event Progressing for config. err = %v", err)
		return
	}

	ee.EmitEvent(config, corev1.EventTypeNormal, ProgressingReason, ProgressingMessage)
}

func (ee eventEmitter) EmitFailingForConfig(reason, message string) {
	config, err := ee.getConfigForEmitter()
	if err != nil {
		log.Printf("Failed to emit event Failing for config. err = %v", err)
		return
	}

	ee.EmitEvent(config, corev1.EventTypeWarning, fmt.Sprintf("%s: %s", FailedReason, reason), fmt.Sprintf("%s: %s", FailedMessage, message))
}

func (ee eventEmitter) EmitAvailableForConfig() {
	config, err := ee.getConfigForEmitter()
	if err != nil {
		log.Printf("Failed to emit event Available for config. err = %v", err)
		return
	}

	ee.EmitEvent(config, corev1.EventTypeNormal, AvailableReason, AvailableMessage)
}

func (ee eventEmitter) EmitModifiedForConfig() {
	config, err := ee.getConfigForEmitter()
	if err != nil {
		log.Printf("Failed to emit event Modified for config. err = %v", err)
		return
	}

	ee.EmitEvent(config, corev1.EventTypeNormal, ModifiedReason, ModifiedMessage)
}

func (ee eventEmitter) getConfigForEmitter() (*cnaov1.NetworkAddonsConfig, error) {
	config := &cnaov1.NetworkAddonsConfig{}
	err := ee.client.Get(context.TODO(), types.NamespacedName{Name: names.OPERATOR_CONFIG}, config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get NetworkAddonsConfig in order to emit event.")
	}

	return config, nil
}
