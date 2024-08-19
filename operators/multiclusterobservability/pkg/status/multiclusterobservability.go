package status

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcov1beta2 "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/v1beta2"
)

const (
	messageReady           = "All components ready"
	messageFailed          = "One or more MultiClusterObservability components failed"
	messagePending         = "One or more MultiClusterObservability components pending on dependencies"
	messageRunning         = "All components are running, but some readiness checks are failing"
	messageMetricsDisabled = "Collect metrics from the managed clusters is disabled"
)

var (
	conditionFailed = metav1.Condition{
		Type:    string(mcov1beta2.ConditionFailed),
		Message: messageFailed,
		Reason:  string(mcov1beta2.ReasonFailedComponents),
	}
	conditionPending = metav1.Condition{
		Type:    string(mcov1beta2.ConditionPending),
		Message: messagePending,
		Reason:  string(mcov1beta2.ReasonPendingComponents),
	}
	conditionRunning = metav1.Condition{
		Type:    string(mcov1beta2.ConditionPending),
		Message: messageRunning,
		Reason:  string(mcov1beta2.ReasonPendingComponents),
	}
	conditionReady = metav1.Condition{
		Type:    string(mcov1beta2.ConditionReady),
		Message: messageReady,
		Reason:  string(mcov1beta2.ReasonReadyComponents),
	}
	conditionMetricsDisabled = metav1.Condition{
		Type:    string(mcov1beta2.ConditionReady),
		Message: messageMetricsDisabled,
		Reason:  string(mcov1beta2.ReasonMetricsDisabled),
	}
)

// DegradedError contains information about why the managed MultiClusterObservability has an invalid configuration.
type DegradedError struct {
	Message string
	Reason  mcov1beta2.MultiClusterObservabilityConditionReason
	Requeue bool
}

func (e *DegradedError) Error() string {
	return fmt.Sprintf("cluster degraded: %s", e.Message)
}

func generateConditions(ctx context.Context, cs *mcov1beta2.MultiClusterObservabilityComponentStatus, k client.Client, stack *mcov1beta2.MultiClusterObservability, degradedErr *DegradedError) ([]metav1.Condition, error) {
	conditions := []metav1.Condition{}

	mainCondition, err := generateCondition(ctx, cs, k, stack, degradedErr)
	if err != nil {
		return nil, err
	}

	conditions = append(conditions, mainCondition)
	return conditions, nil
}

func generateCondition(ctx context.Context, cs *mcov1beta2.MultiClusterObservabilityComponentStatus, k client.Client, stack *mcov1beta2.MultiClusterObservability, degradedErr *DegradedError) (metav1.Condition, error) {
	if degradedErr != nil {
		return metav1.Condition{
			Type:    string(mcov1beta2.ConditionDegraded),
			Message: degradedErr.Message,
			Reason:  string(degradedErr.Reason),
		}, nil
	}

	addonSpec := stack.Spec.ObservabilityAddonSpec
	if addonSpec != nil && !addonSpec.EnableMetrics {
		return conditionMetricsDisabled, nil
	}

	// Check for failed pods first
	failed := len(cs.Grafana[mcov1beta2.PodFailed]) +
		len(cs.ObservatoriumAPI[mcov1beta2.PodFailed]) +
		len(cs.ThanosQuery[mcov1beta2.PodFailed]) +
		len(cs.ThanosQueryFrontend[mcov1beta2.PodFailed]) +
		len(cs.ThanosReceiveController[mcov1beta2.PodFailed]) +
		len(cs.ObservatoriumOperator[mcov1beta2.PodFailed]) +
		len(cs.RBACQueryProxy[mcov1beta2.PodFailed]) +
		len(cs.Alertmanager[mcov1beta2.PodFailed]) +
		len(cs.ThanosCompact[mcov1beta2.PodFailed]) +
		len(cs.ThanosReceive[mcov1beta2.PodFailed]) +
		len(cs.ThanosRule[mcov1beta2.PodFailed]) +
		len(cs.ThanosStoreMemcached[mcov1beta2.PodFailed]) +
		len(cs.ThanosStoreShard[mcov1beta2.PodFailed]) +
		len(cs.MultiClusterObservabilityAddon[mcov1beta2.PodFailed])

	if failed != 0 {
		return conditionFailed, nil
	}

	// Check for pending pods
	pending := len(cs.Grafana[mcov1beta2.PodPending]) +
		len(cs.ObservatoriumAPI[mcov1beta2.PodPending]) +
		len(cs.ThanosQuery[mcov1beta2.PodPending]) +
		len(cs.ThanosQueryFrontend[mcov1beta2.PodPending]) +
		len(cs.ThanosReceiveController[mcov1beta2.PodPending]) +
		len(cs.ObservatoriumOperator[mcov1beta2.PodPending]) +
		len(cs.RBACQueryProxy[mcov1beta2.PodPending]) +
		len(cs.Alertmanager[mcov1beta2.PodPending]) +
		len(cs.ThanosCompact[mcov1beta2.PodPending]) +
		len(cs.ThanosReceive[mcov1beta2.PodPending]) +
		len(cs.ThanosRule[mcov1beta2.PodPending]) +
		len(cs.ThanosStoreMemcached[mcov1beta2.PodPending]) +
		len(cs.ThanosStoreShard[mcov1beta2.PodPending]) +
		len(cs.MultiClusterObservabilityAddon[mcov1beta2.PodPending])

	if pending != 0 {
		return conditionPending, nil
	}

	// Check if there are pods that are running but not ready
	running := len(cs.Grafana[mcov1beta2.PodRunning]) +
		len(cs.ObservatoriumAPI[mcov1beta2.PodRunning]) +
		len(cs.ThanosQuery[mcov1beta2.PodRunning]) +
		len(cs.ThanosQueryFrontend[mcov1beta2.PodRunning]) +
		len(cs.ThanosReceiveController[mcov1beta2.PodRunning]) +
		len(cs.ObservatoriumOperator[mcov1beta2.PodRunning]) +
		len(cs.RBACQueryProxy[mcov1beta2.PodRunning]) +
		len(cs.Alertmanager[mcov1beta2.PodRunning]) +
		len(cs.ThanosCompact[mcov1beta2.PodRunning]) +
		len(cs.ThanosReceive[mcov1beta2.PodRunning]) +
		len(cs.ThanosRule[mcov1beta2.PodRunning]) +
		len(cs.ThanosStoreMemcached[mcov1beta2.PodRunning]) +
		len(cs.ThanosStoreShard[mcov1beta2.PodRunning]) +
		len(cs.MultiClusterObservabilityAddon[mcov1beta2.PodRunning])

	if running > 0 {
		return conditionRunning, nil
	}

	return conditionReady, nil
}
