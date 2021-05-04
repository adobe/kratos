/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package replicas

import (
	"fmt"
	"math"

	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/metrics"
	corev1 "k8s.io/api/core/v1"
)

type ReplicaCalculator struct {
	tolerance float64
}

func NewReplicaCalculator(tolerance float64) *ReplicaCalculator {
	return &ReplicaCalculator{
		tolerance: tolerance,
	}
}

func (rc *ReplicaCalculator) CalculateReplicas(currentReplicas int32, requestedResources map[string]*corev1.ResourceList, scaleMetric v1alpha1.ScaleMetric, metricValues []metrics.MetricValue) (int32, error) {
	metricTarget, err := scaleMetric.GetMetricTarget()
	if err != nil {
		return 0, fmt.Errorf("unknown metric type: %s", scaleMetric.Type)
	}

	switch metricTarget.Type {
	case v1alpha1.ValueMetricType:
		return rc.calculateValue(currentReplicas, metricTarget, metricValues)
	case v1alpha1.AverageValueMetricType:
		return rc.calculateAverageValue(currentReplicas, metricTarget, metricValues)
	case v1alpha1.UtilizationMetricType:
		return rc.calculateUtilization(currentReplicas, scaleMetric, requestedResources, metricValues)
	}
	return 0, fmt.Errorf("replica calculator not implemented yet: %s", metricTarget.Type)
}

func (rc *ReplicaCalculator) calculateUtilization(currentReplicas int32, scaleMetric v1alpha1.ScaleMetric, requestedResources map[string]*corev1.ResourceList, metricValues []metrics.MetricValue) (int32, error) {
	utilization := int64(0)
	for _, metricValue := range metricValues {
		utilization = utilization + metricValue.Value
	}

	totalResources, err := rc.getPodRequestedResource(requestedResources[scaleMetric.Resource.Container], scaleMetric.Resource.Name)

	if err != nil {
		return currentReplicas, err
	}

	if totalResources == 0 {
		return currentReplicas, fmt.Errorf("no resource requests configured for %s", scaleMetric.Resource.Name)
	}

	metricTarget, _ := scaleMetric.GetMetricTarget()
	usageRatio := (float64(utilization) / totalResources) / *metricTarget.AverageUtilization

	if math.Abs(1.0-usageRatio) <= rc.tolerance {
		// return the current replicas if the change would be too small
		return currentReplicas, nil
	}

	replicaCount := int32(math.Ceil(usageRatio * float64(currentReplicas)))

	return replicaCount, nil

}

func (rc *ReplicaCalculator) getPodRequestedResource(resources *corev1.ResourceList, resourceName corev1.ResourceName) (float64, error) {
	switch resourceName {
	case corev1.ResourceCPU:
		return float64(resources.Cpu().Value()), nil
	case corev1.ResourceMemory:
		return float64(resources.Memory().Value()), nil
	case corev1.ResourceStorage:
		return float64(resources.Storage().Value()), nil
	case corev1.ResourceEphemeralStorage:
		return float64(resources.StorageEphemeral().Value()), nil
	case corev1.ResourcePods:
		return float64(resources.Pods().Value()), nil
	}

	return 0, fmt.Errorf("unknown resource metric type: %s", resourceName)
}

func (rc *ReplicaCalculator) calculateValue(currentReplicas int32, metricTarget *v1alpha1.MetricTarget, metricValues []metrics.MetricValue) (int32, error) {
	utilization := int64(0)
	for _, metricValue := range metricValues {
		utilization = utilization + metricValue.Value
	}

	usageRatio := float64(utilization) / float64(metricTarget.Value.Value())

	replicaCount, err := rc.getUsageRatioReplicaCount(currentReplicas, usageRatio)

	if err != nil {
		return 0, err
	}

	return replicaCount, nil
}

func (rc *ReplicaCalculator) calculateAverageValue(currentReplicas int32, metricTarget *v1alpha1.MetricTarget, metricValues []metrics.MetricValue) (replicaCount int32, err error) {
	utilization := int64(0)
	for _, metricValue := range metricValues {
		utilization = utilization + metricValue.Value
	}

	// update number of replicas if the change is large enough
	if currentReplicas != 0 {
		usageRatio := float64(utilization) / (float64(metricTarget.AverageValue.Value()) * float64(currentReplicas))
		if math.Abs(1.0-usageRatio) <= rc.tolerance {
			// return the current replicas if the change would be too small
			return currentReplicas, nil
		}
	}
	replicaCount = int32(math.Ceil(float64(utilization) / float64(metricTarget.AverageValue.Value())))

	return replicaCount, err
}

func (rc *ReplicaCalculator) getUsageRatioReplicaCount(currentReplicas int32, usageRatio float64) (replicaCount int32, err error) {
	if currentReplicas != 0 {
		if math.Abs(1.0-usageRatio) <= rc.tolerance {
			// return the current replicas if the change would be too small
			return currentReplicas, nil
		}
		replicaCount = int32(math.Ceil(usageRatio * float64(currentReplicas)))
	} else {
		// Scale to zero or n pods depending on usageRatio
		replicaCount = int32(math.Ceil(usageRatio))
	}

	return replicaCount, err
}
