// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project
// Licensed under the Apache License 2.0

// Package shared contains shared API Schema definitions for the observability API group
// +kubebuilder:object:generate=true
// +groupName=observability.open-cluster-management.io

package shared

import (
	"net/url"

	corev1 "k8s.io/api/core/v1"
)

// URL is kubebuilder type that validates the containing string is an HTTPS URL.
// +kubebuilder:validation:Pattern=`^https:\/\/`
// +kubebuilder:validation:MaxLength=2083
type URL string

// Validate validates the underlying URL.
func (u URL) Validate() error {
	_, err := url.Parse(string(u))
	return err
}

func (u URL) URL() (*url.URL, error) {
	return url.Parse(string(u))
}

// ObservabilityAddonSpec is the spec of observability addon.
type ObservabilityAddonSpec struct {
	// EnableMetrics indicates the observability addon push metrics to hub server.
	// +optional
	// +kubebuilder:default:=true
	EnableMetrics bool `json:"enableMetrics"`

	// Interval for the observability addon push metrics to hub server.
	// +optional
	// +kubebuilder:default:=300
	// +kubebuilder:validation:Minimum=15
	// +kubebuilder:validation:Maximum=3600
	Interval int32 `json:"interval,omitempty"`

	// Resource requirement for metrics-collector
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type PreConfiguredStorage struct {
	// The key of the secret to select from. Must be a valid secret key.
	// Refer to https://thanos.io/tip/thanos/storage.md/#configuring-access-to-object-storage for a valid content of key.
	// +required
	Key string `json:"key"`
	// Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// +required
	Name string `json:"name"`
	// TLS secret contains the custom certificate for the object store
	// +optional
	TLSSecretName string `json:"tlsSecretName,omitempty"`
	// TLS secret mount path for the custom certificate for the object store
	// +optional
	TLSSecretMountPath string `json:"tlsSecretMountPath,omitempty"`
	// serviceAccountProjection indicates whether mount service account token to thanos pods. Default is false.
	// +optional
	ServiceAccountProjection bool `json:"serviceAccountProjection,omitempty"`
}
