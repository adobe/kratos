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

type behaviorNormalizer struct {
}

func newBehaviorNormalizer() *behaviorNormalizer {
	return &behaviorNormalizer{}
}

func (n *behaviorNormalizer) normalizeReplicas(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus, desiredReplicas int32) int32 {
	maxRecommendation := n.findMaxRecommendation(spec, status, desiredReplicas)

	if maxRecommendation == status.CurrentReplicas {
		return maxRecommendation
	}

	if maxRecommendation > status.CurrentReplicas && spec.Behavior.ScaleUp.SelectPolicy == v1alpha1.DisabledPolicySelect {
		return status.CurrentReplicas
	}

	if maxRecommendation < status.CurrentReplicas && spec.Behavior.ScaleDown.SelectPolicy == v1alpha1.DisabledPolicySelect {
		return status.CurrentReplicas
	}

	if maxRecommendation > status.CurrentReplicas {

		scaleUpLimit := n.calculateScaleUpLimit(spec.Behavior.ScaleUp, status.ScaleUpEvents, status.CurrentReplicas)
		return n.min(common.Max(scaleUpLimit, status.CurrentReplicas), spec.MaxReplicas, maxRecommendation)

	} else if maxRecommendation < status.CurrentReplicas {

		scaleDownLimit := n.calculateScaleDownLimit(spec.Behavior.ScaleDown, status.ScaleDownEvents, status.CurrentReplicas)
		return n.max(common.Min(scaleDownLimit, status.CurrentReplicas), spec.MinReplicas, maxRecommendation)
	}

	return maxRecommendation
}

func (n *behaviorNormalizer) findMaxRecommendation(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus, desiredReplicas int32) int32 {
	if status == nil || status.Recommendations == nil {
		return desiredReplicas
	}

	var stabilizationWondowsSeconds int32
	var recommendationsComparator func(int32, int32) int32

	if desiredReplicas > status.CurrentReplicas {
		stabilizationWondowsSeconds = spec.Behavior.ScaleUp.StabilizationWindowSeconds
		recommendationsComparator = common.Min
	} else {
		stabilizationWondowsSeconds = spec.Behavior.ScaleDown.StabilizationWindowSeconds
		recommendationsComparator = common.Max
	}

	cutOff := time.Now().Add(time.Duration(-stabilizationWondowsSeconds))
	maxRecomendation := desiredReplicas

	for _, recommendation := range status.Recommendations {
		if recommendation.Timestamp.After(cutOff) {
			maxRecomendation = recommendationsComparator(maxRecomendation, recommendation.Replicas)
		}
	}

	return maxRecomendation
}

func (n *behaviorNormalizer) calculateScaleUpLimit(selectRules *v1alpha1.ScaleRules, scaleUpEvents []v1alpha1.ScaleChangeEvent, currentReplicas int32) int32 {
	var result int32 = math.MinInt32
	var proposed int32

	scaleUpSelectFunc := common.Max
	if selectRules.SelectPolicy == v1alpha1.MinPolicySelect {
		scaleUpSelectFunc = common.Min
		result = math.MaxInt32
	}

	for _, policy := range selectRules.Policies {
		replicasAddedInCurrentWindow := n.getReplicasChange(policy.PeriodSeconds, scaleUpEvents)
		windowStartReplicas := currentReplicas - replicasAddedInCurrentWindow
		if policy.Type == v1alpha1.PodsScalingPolicy {
			proposed = int32(windowStartReplicas + policy.Value)
		} else if policy.Type == v1alpha1.PercentScalingPolicy {
			// the proposal has to be rounded up because the proposed change might not increase the replica count causing the target to never scale up
			proposed = int32(math.Ceil(float64(windowStartReplicas) * (1 + float64(policy.Value)/100)))
		}
		result = scaleUpSelectFunc(result, proposed)
	}

	return result
}

func (n *behaviorNormalizer) calculateScaleDownLimit(scaleRules *v1alpha1.ScaleRules, scaleDownEvents []v1alpha1.ScaleChangeEvent, currentReplicas int32) int32 {
	var result int32 = math.MaxInt32
	var proposed int32

	scaleDownSelectFunc := common.Min

	if scaleRules.SelectPolicy == v1alpha1.MinPolicySelect {
		scaleDownSelectFunc = common.Max
		result = math.MinInt32
	}

	for _, policy := range scaleRules.Policies {
		replicasDeletedInCurrentWindow := n.getReplicasChange(policy.PeriodSeconds, scaleDownEvents)
		windowStartReplicas := currentReplicas + replicasDeletedInCurrentWindow
		if policy.Type == v1alpha1.PodsScalingPolicy {
			proposed = windowStartReplicas - policy.Value
		} else if policy.Type == v1alpha1.PercentScalingPolicy {
			proposed = int32(float64(windowStartReplicas) * (1 - float64(policy.Value)/100))
		}
		result = scaleDownSelectFunc(result, proposed)
	}
	return result

}

func (n *behaviorNormalizer) getReplicasChange(windowSeconds int32, scaleEvents []v1alpha1.ScaleChangeEvent) int32 {
	windowAsDuration := time.Second * time.Duration(windowSeconds)
	cutOff := time.Now().Add(-windowAsDuration)

	totalReplicasChange := int32(0)

	for _, event := range scaleEvents {
		if event.Timestamp.After(cutOff) {
			totalReplicasChange += event.ReplicaChange
		}
	}

	return totalReplicasChange
}

func (n *behaviorNormalizer) min(nums ...int32) int32 {
	result := int32(math.MaxInt32)

	for _, value := range nums {
		result = common.Min(result, value)
	}

	return result
}

func (n *behaviorNormalizer) max(nums ...int32) int32 {
	result := int32(math.MinInt32)

	for _, value := range nums {
		result = common.Max(result, value)
	}

	return result
}
