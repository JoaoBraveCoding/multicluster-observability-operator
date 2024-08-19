package status

import (
	"context"

	"github.com/ViaQ/logerr/v2/kverrors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcov1beta2 "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/v1beta2"
	"github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/pkg/config"
)

// generateComponentStatus updates the pod status map component
func generateComponentStatus(ctx context.Context, c client.Client, instance *mcov1beta2.MultiClusterObservability) (*mcov1beta2.MultiClusterObservabilityComponentStatus, error) {
	var err error
	result := &mcov1beta2.MultiClusterObservabilityComponentStatus{}
	result.MultiClusterObservabilityAddon, err = appendPodStatus(ctx, c, config.MultiClusterObservabilityAddon, instance.Name)
	if err != nil {
		return nil, kverrors.Wrap(err, "failed lookup MultiClusterObservability component pods status", "name", config.MultiClusterObservabilityAddon)
	}

	return result, nil
}

func appendPodStatus(ctx context.Context, c client.Client, component, instanceName string) (mcov1beta2.PodStatusMap, error) {
	psm := mcov1beta2.PodStatusMap{}
	pods := &corev1.PodList{}
	opts := []client.ListOption{
		client.MatchingLabels(config.ComponentLabels(component, instanceName)),
		client.InNamespace(config.GetDefaultNamespace()),
	}
	if err := c.List(ctx, pods, opts...); err != nil {
		return nil, kverrors.Wrap(err, "failed to list pods for MultiClusterObservability component", "name", instanceName, "component", component)
	}
	for _, pod := range pods.Items {
		status := podStatus(&pod)
		psm[status] = append(psm[status], pod.Name)
	}
	return psm, nil
}

func podStatus(pod *corev1.Pod) mcov1beta2.PodStatus {
	status := pod.Status
	switch status.Phase {
	case corev1.PodFailed:
		return mcov1beta2.PodFailed
	case corev1.PodPending:
		return mcov1beta2.PodPending
	case corev1.PodRunning:
	default:
		return mcov1beta2.PodStatusUnknown
	}

	for _, c := range status.ContainerStatuses {
		if !c.Ready {
			return mcov1beta2.PodRunning
		}
	}

	return mcov1beta2.PodReady
}
