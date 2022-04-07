package statusmanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cnao "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/shared"
	cnaov1 "github.com/kubevirt/cluster-network-addons-operator/pkg/apis/networkaddonsoperator/v1"
	eventemitter "github.com/kubevirt/cluster-network-addons-operator/pkg/eventemitter"
)

const (
	conditionsUpdateRetries  = 10
	conditionsUpdateCoolDown = 50 * time.Millisecond
)

var (
	operatorVersion string
)

func init() {
	operatorVersion = os.Getenv("OPERATOR_VERSION")
}

// StatusLevel is used to sort priority of reported failure conditions. When operator is failing
// on two levels (e.g. handling configuration and deploying pods), only the higher level will be
// reported. This is needed, so failing pods from previous run won't silence failing render of
// the current run.
type StatusLevel int

const (
	OperatorConfig StatusLevel = iota
	PodDeployment  StatusLevel = iota
	maxStatusLevel StatusLevel = iota
)

// StatusManager coordinates changes to NetworkAddonsConfig.Status
type StatusManager struct {
	client client.Client
	name   string

	failing [maxStatusLevel]*conditionsv1.Condition

	generation  int64
	daemonSets  []types.NamespacedName
	deployments []types.NamespacedName

	containers   []cnao.Container
	mux          sync.Mutex
	eventEmitter eventemitter.EventEmitter
}

func New(mgr manager.Manager, name string) *StatusManager {
	return &StatusManager{
		client:       mgr.GetClient(),
		name:         name,
		eventEmitter: eventemitter.New(mgr),
	}
}

// Set updates the NetworkAddonsConfig.Status with the provided conditions.
// Since Update call can fail due to a collision with someone else writing into
// the status, calling set is tried several times.
// current collision problem is detected by functional tests
func (status *StatusManager) Set(reachedAvailableLevel bool, conditions ...conditionsv1.Condition) {
	for i := 0; i < conditionsUpdateRetries; i++ {
		err := status.set(reachedAvailableLevel, conditions...)
		if err == nil {
			log.Print("Successfully updated status conditions")
			return
		}
		log.Printf("Failed calling status Set %d/%d: %v", i+1, conditionsUpdateRetries, err)
		time.Sleep(conditionsUpdateCoolDown)
	}
	log.Print("Failed to update conditions within given number of retries")
}

// set updates the NetworkAddonsConfig.Status with the provided conditions
func (status *StatusManager) set(reachedAvailableLevel bool, conditions ...conditionsv1.Condition) error {
	config, err := status.getCurrentNetworkAddonsConfig()
	if err != nil {
		log.Printf("Failed to get NetworkAddonsOperator %q in order to update its State: %v", status.name, err)
		return nil
	}

	patch := client.MergeFrom(config.DeepCopy())
	oldStatus := config.Status.DeepCopy()

	// Update Status field with given conditions
	for _, condition := range conditions {
		conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions, condition)
	}

	// Glue condition logic together
	if status.failing[OperatorConfig] != nil &&
		!conditionsv1.IsStatusConditionPresentAndEqual(config.Status.Conditions, conditionsv1.ConditionAvailable, corev1.ConditionFalse) {
		// In case the operator is failing, we should not report it as being ready. This has
		// to be done even when the operator is running fine based on the previous configuration
		// and the only failing thing is validation of new config.
		reason := "Failing"
		message := "Unable to apply desired configuration"
		status.eventEmitter.EmitFailingForConfig(reason, message)
		conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
			conditionsv1.Condition{
				Type:    conditionsv1.ConditionAvailable,
				Status:  corev1.ConditionFalse,
				Reason:  reason,
				Message: message,
			},
		)

		// Implicitly mark as not Progressing, that indicates that human interaction is needed
		conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
			conditionsv1.Condition{
				Type:    conditionsv1.ConditionProgressing,
				Status:  corev1.ConditionFalse,
				Reason:  "InvalidConfiguration",
				Message: "Human interaction is needed, please fix the desired configuration",
			},
		)
	} else if status.failing[PodDeployment] != nil &&
		!conditionsv1.IsStatusConditionPresentAndEqual(config.Status.Conditions, conditionsv1.ConditionAvailable, corev1.ConditionFalse) {
		// In case pod deployment is in progress, implicitly mark as not Available
		reason := "Failing"
		message := "Some problems occurred while deploying components' pods"
		status.eventEmitter.EmitFailingForConfig(reason, message)
		conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
			conditionsv1.Condition{
				Type:    conditionsv1.ConditionAvailable,
				Status:  corev1.ConditionFalse,
				Reason:  reason,
				Message: message,
			},
		)
	} else if conditionsv1.IsStatusConditionPresentAndEqual(config.Status.Conditions, conditionsv1.ConditionProgressing, corev1.ConditionTrue) {
		// In case that the status field has been updated with "Progressing" condition, make sure that
		// "Ready" condition is set to False, even when not explicitly set.
		conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
			conditionsv1.Condition{
				Type:    conditionsv1.ConditionAvailable,
				Status:  corev1.ConditionFalse,
				Reason:  "Startup",
				Message: "Configuration is in process",
			},
		)
	} else if reachedAvailableLevel {
		if config.Spec.NMState != nil {
			// CNAO doesn't support nmstate deployment anymore, set Degraded state is nmstate is requested in NetworkAddonsConfig
			reason := "InvalidConfiguration"
			message := "NMState deployment is not supported by CNAO anymore, please install Kubernetes NMState Operator"
			status.eventEmitter.EmitFailingForConfig(reason, message)
			conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
				conditionsv1.Condition{
					Type:    conditionsv1.ConditionDegraded,
					Status:  corev1.ConditionTrue,
					Reason:  reason,
					Message: message,
				},
			)
		} else {
			// If successfully deployed all components and is not failing on anything, mark as Available
			status.eventEmitter.EmitAvailableForConfig()
			conditionsv1.SetStatusConditionNoHeartbeat(&config.Status.Conditions,
				conditionsv1.Condition{
					Type:   conditionsv1.ConditionAvailable,
					Status: corev1.ConditionTrue,
				},
			)
			config.Status.ObservedVersion = operatorVersion
		}
	}

	// Make sure to expose deployed containers
	_, _, config.Status.Containers, status.generation = status.GetAttributes()

	// Expose currently handled version
	config.Status.OperatorVersion = operatorVersion
	config.Status.TargetVersion = operatorVersion

	// Failing condition had been replaced by Degraded in 0.12.0, drop it from CR if needed
	conditionsv1.RemoveStatusCondition(&config.Status.Conditions, conditionsv1.ConditionType("Failing"))

	if (*oldStatus).DeepEqual(config.Status) {
		return nil
	}

	// Patch NetworkAddonsConfig's status
	err = status.client.Status().Patch(context.TODO(), config, patch)
	if err != nil {
		return fmt.Errorf("Failed to patch NetworkAddonsConfig %q Status: %v", config.Name, err)
	}
	return nil
}

func (status *StatusManager) getCurrentNetworkAddonsConfig() (*cnaov1.NetworkAddonsConfig, error) {
	// Read the current NetworkAddonsConfig
	config := &cnaov1.NetworkAddonsConfig{ObjectMeta: metav1.ObjectMeta{Name: status.name}}
	err := status.client.Get(context.TODO(), types.NamespacedName{Name: status.name}, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// IsStatusAvailable returns true if NetworkAddonsConfig intance is in Available True state
func (status *StatusManager) IsStatusAvailable() bool {
	config, err := status.getCurrentNetworkAddonsConfig()
	if err != nil {
		log.Printf("Failed to get NetworkAddonsOperator %q in order to assess availability: %v", status.name, err)
		return false
	}

	for _, condition := range config.Status.Conditions {
		if condition.Type == conditionsv1.ConditionAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

// getFailureStateCondition returns a general state condition of the system. if any of the status.Failing statuses
// are on, returns the highest priority, otherwise return Degraded=False condition
func (status *StatusManager) getFailureStateCondition() conditionsv1.Condition {
	for _, c := range status.failing {
		if c != nil {
			return *c
		}
	}

	return conditionsv1.Condition{
		Type:   conditionsv1.ConditionDegraded,
		Status: corev1.ConditionFalse,
	}
}

// SetFailing marks the operator as Failing with the given reason and message. If it
// is not already failing for a lower-level reason, the operator's status will be updated.
func (status *StatusManager) SetFailing(level StatusLevel, reason, message string) {
	status.eventEmitter.EmitFailingForConfig(reason, message)
	status.failing[level] = &conditionsv1.Condition{
		Type:    conditionsv1.ConditionDegraded,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	}
	reachedAvailableLevel := false
	status.Set(reachedAvailableLevel, status.getFailureStateCondition())
}

// MarkStatusLevelNotFailing marks the operator as not Failing at the given level.
func (status *StatusManager) MarkStatusLevelNotFailing(level StatusLevel) {
	if status.failing[level] != nil {
		status.failing[level] = nil
	}
}

func (status *StatusManager) SetAttributes(daemonSets []types.NamespacedName, deployments []types.NamespacedName, containers []cnao.Container, generation int64) {
	status.mux.Lock()
	defer status.mux.Unlock()
	status.daemonSets = daemonSets
	status.deployments = deployments
	status.containers = containers
	status.generation = generation
}

func (status *StatusManager) GetAttributes() ([]types.NamespacedName, []types.NamespacedName, []cnao.Container, int64) {
	status.mux.Lock()
	defer status.mux.Unlock()
	return status.daemonSets, status.deployments, status.containers, status.generation
}

// SetFromOperator sets the operator status
func (status *StatusManager) SetFromOperator() {
	conditions := []conditionsv1.Condition{}
	conditions = append(conditions, status.getFailureStateCondition())
	// Available state is set to true when all tracked objects are ready, and thus is not
	// related to the status of the operator itself
	availableStatusReached := false

	// set the aggregated conditions to status-manager
	status.Set(availableStatusReached, conditions...)
}

// SetFromPods sets the operator status to Failing, Progressing, or Available, based on
// the current status of the manager's DaemonSets and Deployments. However, this is a
// no-op if the StatusManager is currently marked as failing due to a configuration error.
func (status *StatusManager) SetFromPods() {
	progressing := []string{}
	daemonSets, deployments, _, generation := status.GetAttributes()

	// Iterate all owned DaemonSets and check whether they are progressing smoothly or have been
	// already deployed.
	for _, dsName := range daemonSets {
		// First check whether DaemonSet namespace exists
		ns := &corev1.Namespace{}
		if err := status.client.Get(context.TODO(), types.NamespacedName{Name: dsName.Namespace}, ns); err != nil {
			if errors.IsNotFound(err) {
				status.SetFailing(PodDeployment, "NoNamespace",
					fmt.Sprintf("Namespace %q does not exist", dsName.Namespace))
			} else {
				status.SetFailing(PodDeployment, "InternalError",
					fmt.Sprintf("Internal error deploying pods: %v", err))
			}
			return
		}

		// Then check whether is the DaemonSet created on Kubernetes API server
		ds := &appsv1.DaemonSet{}
		if err := status.client.Get(context.TODO(), dsName, ds); err != nil {
			if errors.IsNotFound(err) {
				status.SetFailing(PodDeployment, "NoDaemonSet",
					fmt.Sprintf("Expected DaemonSet %q does not exist", dsName.String()))
			} else {
				status.SetFailing(PodDeployment, "InternalError",
					fmt.Sprintf("Internal error deploying pods: %v", err))
			}
			return
		}

		// Finally check whether Pods belonging to this DaemonSets are being started or they
		// are being scheduled.
		if ds.Status.NumberUnavailable > 0 {
			progressing = append(progressing, fmt.Sprintf("DaemonSet %q is not available (awaiting %d nodes)", dsName.String(), ds.Status.NumberUnavailable))
		} else if ds.Status.NumberAvailable == 0 && ds.Status.DesiredNumberScheduled != 0 {
			progressing = append(progressing, fmt.Sprintf("DaemonSet %q is not yet scheduled on any nodes", dsName.String()))
		} else if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			progressing = append(progressing, fmt.Sprintf("DaemonSet %q update is rolling out (%d out of %d updated)", dsName.String(), ds.Status.UpdatedNumberScheduled, ds.Status.DesiredNumberScheduled))
		} else if ds.Generation > ds.Status.ObservedGeneration {
			progressing = append(progressing, fmt.Sprintf("DaemonSet %q update is being processed (generation %d, observed generation %d)", dsName.String(), ds.Generation, ds.Status.ObservedGeneration))
		}
	}

	// Do the same for Deployments. Iterate all owned Deployments and check whether they are
	// progressing smoothly or have been already deployed.
	for _, depName := range deployments {
		// First check whether Deployment namespace exists
		ns := &corev1.Namespace{}
		if err := status.client.Get(context.TODO(), types.NamespacedName{Name: depName.Namespace}, ns); err != nil {
			if errors.IsNotFound(err) {
				status.SetFailing(PodDeployment, "NoNamespace",
					fmt.Sprintf("Namespace %q does not exist", depName.Namespace))
			} else {
				status.SetFailing(PodDeployment, "InternalError",
					fmt.Sprintf("Internal error deploying pods: %v", err))
			}
			return
		}

		// Then check whether is the Deployment created on Kubernetes API server
		dep := &appsv1.Deployment{}
		if err := status.client.Get(context.TODO(), depName, dep); err != nil {
			if errors.IsNotFound(err) {
				status.SetFailing(PodDeployment, "NoDeployment",
					fmt.Sprintf("Expected Deployment %q does not exist", depName.String()))
			} else {
				status.SetFailing(PodDeployment, "InternalError",
					fmt.Sprintf("Internal error deploying pods: %v", err))
			}
			return
		}

		// Finally check whether Pods belonging to this Deployments are being started or they
		// are being scheduled.
		if dep.Status.UnavailableReplicas > 0 {
			progressing = append(progressing, fmt.Sprintf("Deployment %q is not available (awaiting %d nodes)", depName.String(), dep.Status.UnavailableReplicas))
		} else if dep.Status.AvailableReplicas == 0 {
			progressing = append(progressing, fmt.Sprintf("Deployment %q is not yet scheduled on any nodes", depName.String()))
		} else if dep.Status.ObservedGeneration < dep.Generation {
			progressing = append(progressing, fmt.Sprintf("Deployment %q update is being processed (generation %d, observed generation %d)", depName.String(), dep.Generation, dep.Status.ObservedGeneration))
		}
	}

	// aggregate non-failing conditions
	conditions := []conditionsv1.Condition{}
	availableStatusReached := false

	// If there are any progressing Pods, list them in the condition with their state. Otherwise,
	// mark Progressing condition as False.
	if len(progressing) > 0 {
		status.eventEmitter.EmitProgressingForConfig()
		conditions = append(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionProgressing,
			Status:  corev1.ConditionTrue,
			Reason:  "Deploying",
			Message: strings.Join(progressing, "\n"),
		})
	} else {
		conditions = append(conditions, conditionsv1.Condition{
			Type:   conditionsv1.ConditionProgressing,
			Status: corev1.ConditionFalse,
		})
	}

	// If all pods are being created, mark deployment as not failing
	status.MarkStatusLevelNotFailing(PodDeployment)
	conditions = append(conditions, status.getFailureStateCondition())

	config, err := status.getCurrentNetworkAddonsConfig()
	if err != nil {
		return
	}

	// if generation hasn't changed and all containers are deployed, mark as Available
	if config.GetGeneration() == generation && len(progressing) == 0 {
		availableStatusReached = true
	}

	// set the aggregated conditions to status-manager
	status.Set(availableStatusReached, conditions...)
}
