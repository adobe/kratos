/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package normalizer

import (
	"math"
	"time"

	"github.com/adobe/kratos/pkg/api/common"
	"github.com/adobe/kratos/pkg/api/v1alpha1"
)

const (
	scaleUpLimitFactor  = 2.0
	scaleUpLimitMinimum = 4.0
)

type standardNormalizer struct {
}

func newStandardNormalizer() *standardNormalizer {
	return &standardNormalizer{}
}

func (n *standardNormalizer) normalizeReplicas(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus, desiredReplicas int32) int32 {
	maxRecommendation := n.findMaxRecommendation(spec.StabilizationWindowSeconds, status, desiredReplicas)

	scaleUpLimit := n.calculateScaleUpLimit(status.CurrentReplicas)
	maxAllowedScaleReplicas := common.Min(spec.MaxReplicas, scaleUpLimit)

	if maxRecommendation > maxAllowedScaleReplicas {
		return maxAllowedScaleReplicas
	}

	if maxRecommendation < spec.MinReplicas {
		return spec.MinReplicas
	}

	return maxRecommendation
}

func (n *standardNormalizer) findMaxRecommendation(stabilizationWindowSec int32, status *v1alpha1.KratosStatus, desiredReplicas int32) int32 {
	if status == nil || status.Recommendations == nil {
		return desiredReplicas
	}

	cutOff := time.Now().Add(time.Duration(-stabilizationWindowSec))
	maxRecommendation := desiredReplicas
	for _, recommendation := range status.Recommendations {
		if maxRecommendation < recommendation.Replicas && recommendation.Timestamp.After(cutOff) {
			maxRecommendation = recommendation.Replicas
		}
	}

	return maxRecommendation
}

func (n *standardNormalizer) calculateScaleUpLimit(currentReplicas int32) int32 {
	return int32(math.Max(scaleUpLimitFactor*float64(currentReplicas), scaleUpLimitMinimum))
}
