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
	"errors"
	"fmt"
	"math"

	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/metrics"
)

type ReplicaCalculator struct {
	tolerance float64
}

func NewReplicaCalculator(tolerance float64) *ReplicaCalculator {
	return &ReplicaCalculator{
		tolerance: tolerance,
	}
}

func (rc *ReplicaCalculator) CalculateReplicas(currentReplicas int32, metricTarget v1alpha1.MetricTarget, metricValues []metrics.MetricValue) (int32, error) {
	switch metricTarget.Type {
	case v1alpha1.ValueMetricType:
		return rc.calculateValue(currentReplicas, metricTarget, metricValues)
	case v1alpha1.AverageValueMetricType:
		return rc.calculateAverageValue(currentReplicas, metricTarget, metricValues)
	}
	return 0, errors.New(fmt.Sprintf("Replica calculator not implemented yet: %s", metricTarget.Type))
}

func (rc *ReplicaCalculator) calculateValue(currentReplicas int32, metricTarget v1alpha1.MetricTarget, metricValues []metrics.MetricValue) (int32, error) {
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

func (rc *ReplicaCalculator) calculateAverageValue(currentReplicas int32, metricTarget v1alpha1.MetricTarget, metricValues []metrics.MetricValue) (replicaCount int32, err error) {
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
