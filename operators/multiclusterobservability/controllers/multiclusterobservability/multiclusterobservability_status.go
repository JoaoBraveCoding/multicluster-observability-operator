// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project
// Licensed under the Apache License 2.0

package multiclusterobservability

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcov1beta2 "github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/api/v1beta2"
	"github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/pkg/config"
	"github.com/stolostron/multicluster-observability-operator/operators/multiclusterobservability/pkg/status"
)

func checkReadyStatus(ctx context.Context, c client.Client, mco *mcov1beta2.MultiClusterObservability) error {
	if err := mcoaCRDsInstalled(c, mco); err != nil {
		return err
	}

	if err := checkObjStorageStatus(ctx, c, mco); err != nil {
		return err
	}

	return nil
}

func mcoaCRDsInstalled(c client.Client, mco *mcov1beta2.MultiClusterObservability) error {
	if mco.Spec.Capabilities == nil {
		return nil
	}

	if mco.Spec.Capabilities.Platform == nil && mco.Spec.Capabilities.UserWorkloads == nil {
		return nil
	}

	var missing []string

outer:
	for _, crdName := range config.GetMCOASupportedCRDNames() {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		key := client.ObjectKey{Name: crdName}

		err := c.Get(context.TODO(), key, crd)
		if client.IgnoreAlreadyExists(err) != nil {
			missing = append(missing, crdName)
			continue
		}

		version := config.GetMCOASupportedCRDVersion(crdName)

		for _, crdVersion := range crd.Spec.Versions {
			if crdVersion.Name == version && crdVersion.Served {
				continue outer
			}
		}

		missing = append(missing, crdName)
	}

	if len(missing) == 0 {
		return nil
	}
	tmpl := "MultiCluster-Observability-Addon degraded because the following CRDs are not installed on the hub: %s"

	var missingVersions []string
	for _, name := range missing {
		version := config.GetMCOASupportedCRDVersion(name)
		missingVersions = append(missingVersions, fmt.Sprintf("%s(%s)", name, version))
	}

	msg := fmt.Sprintf(tmpl, strings.Join(missingVersions, ", "))

	return &status.DegradedError{
		Reason:  mcov1beta2.ReasonMCOAMissingCRDs,
		Message: msg,
	}
}

func checkObjStorageStatus(
	ctx context.Context,
	c client.Client,
	mco *mcov1beta2.MultiClusterObservability,
) error {
	objStorageConf := mco.Spec.StorageConfig.MetricObjectStorage
	secret := &corev1.Secret{}
	namespacedName := types.NamespacedName{
		Name:      objStorageConf.Name,
		Namespace: config.GetDefaultNamespace(),
	}

	if err := c.Get(ctx, namespacedName, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return &status.DegradedError{
				Reason:  mcov1beta2.ReasonMissingObjectStorageSecret,
				Message: "Missing object storage secret",
				Requeue: false,
			}
		}
		return fmt.Errorf("failed to lookup object storage secret: %w", err)
	}

	if err := config.CheckObjStorageConf(objStorageConf.Key, secret); err != nil {
		return &status.DegradedError{
			Message: fmt.Sprintf("Invalid object storage secret contents: %s", err),
			Reason:  mcov1beta2.ReasonInvalidObjectStorageSecret,
			Requeue: false,
		}
	}

	return nil
}
