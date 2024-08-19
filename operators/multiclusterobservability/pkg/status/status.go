package status

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcov1beta2 "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/v1beta2"
)

func RefreshStatus(ctx context.Context, c client.Client, req ctrl.Request, now time.Time, degradedErr *DegradedError) error {
	instance := &mcov1beta2.MultiClusterObservability{}
	err := c.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to lookup instance %s: %w", instance.Name, err)
	}

	cs, err := generateComponentStatus(ctx, c, instance)
	if err != nil {
		return err
	}
	activeConditions, err := generateConditions(ctx, cs, c, instance, degradedErr)
	if err != nil {
		return err
	}

	metaTime := metav1.NewTime(now)
	for _, c := range activeConditions {
		c.LastTransitionTime = metaTime
		c.Status = metav1.ConditionTrue
	}

	statusUpdater := func(instance *mcov1beta2.MultiClusterObservability) {
		instance.Status.Components = *cs
		instance.Status.Conditions = mergeConditions(instance.Status.Conditions, activeConditions, metaTime)
	}

	statusUpdater(instance)
	err = c.Status().Update(ctx, instance)
	switch {
	case err == nil:
		return nil
	case apierrors.IsConflict(err):
		// break into retry-logic below on conflict
		break
	default:
		// return non-conflict errors
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := c.Get(ctx, req.NamespacedName, instance); err != nil {
			return err
		}

		statusUpdater(instance)
		return c.Status().Update(ctx, instance)
	})
}
