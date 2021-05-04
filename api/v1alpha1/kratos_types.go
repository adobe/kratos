/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KratosSpec defines the desired state of Kratos
type KratosSpec struct {
	Target ScaleTargetReference `json:"target" protobuf:"bytes,1,opt,name=target"`

	Algorithm Algorithm `json:"algorithm" protobuf:"bytes,2,opt,name=algorithm"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.
	MinReplicas int32 `json:"minReplicas,omitempty" protobuf:"varint,3,opt,name=minReplicas"`

	// upper limit for the number of pods that can be set by the autoscaler; cannot be smaller than MinReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,4,opt,name=maxReplicas"`

	// stabilization window in seconds
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds" protobuf:"varint,5,opt,name=stabilizationWindowSeconds"`

	// Metrics to use for scaling
	Metrics []ScaleMetric `json:"metrics,omitempty" protobuf:"bytes,6,rep,name=metrics"`

	// Up or Down scaling behavior
	Behavior *ScaleBehavior `json:"behavior,omitempty" protobuf:"bytes,7,opt,name=behavior"`
}

// KratosStatus defines the observed state of Kratos
type KratosStatus struct {

	//Time when stabilization window ends
	StabilizationEndTime *metav1.Time `json:"stabilizationEndTime,omitempty" protobuf:"bytes,1,opt,name=stabilizationEndTime"`

	//current target replicas
	CurrentReplicas int32 `json:"currentReplicas" protobuf:"varint,2,opt,name=currentReplicas"`

	//desired number of replicas for target
	DesiredReplicas int32 `json:"desiredReplicas" protobuf:"varint,3,opt,name=desiredReplicas"`

	//scale recommendations
	Recommendations []Recommendation `json:"recommendations" protobuf:"varint,4,opt,name=recommendations"`

	//scale up events
	ScaleUpEvents []ScaleChangeEvent `json:"scaleUpEvents" protobuf:"varint,5,opt,name=scaleUpEvents"`

	//scale down events
	ScaleDownEvents []ScaleChangeEvent `json:"scaleDownEvents" protobuf:"varint,6,opt,name=scaleDownEvents"`
}

// ScalingTargetReference identifies target to scale
type ScaleTargetReference struct {
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}

// Algorithm to use for scale
type Algorithm struct {
	Type    string            `json:"type" protobuf:"bytes,1,opt,name=type"`
	Options map[string]string `json:"options,omitempty" protobuf:"bytes,1,opt,name=options"`
}

type MetricType string

const (
	ResourceScaleMetricType   MetricType = "Resource"
	PodScaleMetricType        MetricType = "Pod"
	ObjectScaleMetricType     MetricType = "Object"
	ExternalScaleMetricType   MetricType = "External"
	PrometheusScaleMetricType MetricType = "Prometheus"
)

type ScaleMetric struct {
	Type MetricType `json:"type" protobuf:"bytes,1,name=type"`

	// resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing each pod in the
	// current scale target (e.g. CPU or memory). Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics using the "pods" source.
	// +optional
	Resource *ResourceMetricSource `json:"resource,omitempty" protobuf:"bytes,1,opt,name=resource"`

	// pods refers to a metric describing each pod in the current scale target
	// (for example, transactions-processed-per-second).  The values will be
	// averaged together before being compared to the target value.
	// +optional
	Pods *PodsMetricSource `json:"pods,omitempty" protobuf:"bytes,2,opt,name=pods"`

	// object refers to a metric describing a single kubernetes object
	// (for example, hits-per-second on an Ingress object).
	// +optional
	Object *ObjectMetricSource `json:"object,omitempty" protobuf:"bytes,3,opt,name=object"`

	// external refers to a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	// +optional
	External *ExternalMetricSource `json:"external,omitempty" protobuf:"bytes,5,opt,name=external"`

	Prometheus *PrometheusMetricSource `json:"prometheus,omitempty" protobuf:"bytes,6,opt,name=prometheus"`
}

// ResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ResourceMetricSource struct {
	// name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// container is the name of the container in the pods of the scaling target.
	// +optional
	Container string `json:"container" protobuf:"bytes,3,opt,name=container"`
}

// PodsMetricSource indicates how to scale on a metric describing each pod in
// the current scale target (for example, transactions-processed-per-second).
// The values will be averaged together before being compared to the target
// value.
type PodsMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// ObjectMetricSource indicates how to scale on a metric describing a
// kubernetes object (for example, hits-per-second on an Ingress object).
type ObjectMetricSource struct {
	DescribedObject ScaleTargetReference `json:"describedObject" protobuf:"bytes,1,name=describedObject"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,3,name=metric"`
}

// ExternalMetricSource indicates how to scale on a metric not associated with
// any Kubernetes object (for example length of queue in cloud
// messaging service, or QPS from loadbalancer running outside of cluster).
type ExternalMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

type PrometheusMetricSource struct {

	// Metrics query in PromQL language. Must return single value.
	MetricQuery string `json:"metricQuery" protobuf:"bytes,1,name=metricQuery"`

	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`

	//Prometheus endpoint for retrieving metrics. Default to global setting set at the Operator level
	PrometheusEndpoint string `json:"prometheusEndpoint" protobuf:"bytes,3,name=prometheusEndpoint"`
}

// MetricIdentifier defines the name and optionally selector for a metric
type MetricIdentifier struct {
	// name is the name of the given metric
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// selector is the string-encoded form of a standard kubernetes label selector for the given metric
	// When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping.
	// When unset, just the metricName will be used to gather metrics.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,name=selector"`
}

// MetricTargetType specifies the type of metric being targeted, and should be either
// "Value", "AverageValue", or "Utilization"
type MetricTargetType string

const (
	// UtilizationMetricType declares a MetricTarget is an AverageUtilization value
	UtilizationMetricType MetricTargetType = "Utilization"
	// ValueMetricType declares a MetricTarget is a raw value
	ValueMetricType MetricTargetType = "Value"
	// AverageValueMetricType declares a MetricTarget is an average across all relevant pods (as a quantity)
	AverageValueMetricType MetricTargetType = "AverageValue"
)

// MetricTarget defines the target value, average value, or average utilization of a specific metric
type MetricTarget struct {
	// type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type" protobuf:"bytes,1,name=type"`
	// value is the target value of the metric (as a quantity).
	// +optional
	Value *resource.Quantity `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// averageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *resource.Quantity `json:"averageValue,omitempty" protobuf:"bytes,3,opt,name=averageValue"`
	// averageUtilization is the target value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// Currently only valid for Resource metric source type
	// +optional
	AverageUtilization *float64 `json:"averageUtilization,omitempty" protobuf:"bytes,4,opt,name=averageUtilization"`
}

type ScaleBehavior struct {
	ScaleUp   *ScaleRules `json:"scaleUp,omitempty" protobuf:"bytes,1,opt,name=scaleUp"`
	ScaleDown *ScaleRules `json:"scaleDown,omitempty" protobuf:"bytes,2,opt,name=scaleDown"`
}

type ScaleRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds" protobuf:"varint,1,opt,name=stabilizationWindowSeconds"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value MaxPolicySelect is used.
	// +optional
	SelectPolicy ScalingPolicySelect `json:"selectPolicy,omitempty" protobuf:"bytes,2,opt,name=selectPolicy"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +optional
	Policies []ScalingPolicy `json:"policies,omitempty" protobuf:"bytes,3,rep,name=policies"`
}

// ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction
type ScalingPolicySelect string

const (
	// MaxPolicySelect selects the policy with the highest possible change.
	MaxPolicySelect ScalingPolicySelect = "Max"
	// MinPolicySelect selects the policy with the lowest possible change.
	MinPolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// ScalingPolicyType is the type of the policy which could be used while making scaling decisions.
type ScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy ScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy ScalingPolicyType = "Percent"
)

// ScalingPolicy is a single policy which must hold true for a specified past interval.
type ScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type ScalingPolicyType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ScalingPolicyType"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int32 `json:"value" protobuf:"varint,2,opt,name=value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int32 `json:"periodSeconds" protobuf:"varint,3,opt,name=periodSeconds"`
}

// Recommendation details
type Recommendation struct {
	// timestamp of the recommendation
	Timestamp metav1.Time `json:"timestamp" protobuf:"varint,1,opt,name=timestamp"`
	// recommended replicas
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`
}

// ScaleChangeEvent holds timestamp and replica delta
type ScaleChangeEvent struct {
	// timestamp of the event
	Timestamp metav1.Time `json:"timestamp" protobuf:"varint,1,opt,name=timestamp"`
	// change of replicas
	ReplicaChange int32 `json:"replicaChange" protobuf:"varint,2,opt,name=replicaChange"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Kratos is the Schema for the kratos API
type Kratos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KratosSpec   `json:"spec,omitempty"`
	Status KratosStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KratosList contains a list of Kratos
type KratosList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kratos `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kratos{}, &KratosList{})
}
