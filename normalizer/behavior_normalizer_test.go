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
	"github.com/adobe/kratos/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("BehaviorNormalizer", func() {
	DescribeTable("ScaleUp disabled, ScaleDown disabled",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := createSpec(minReplicas, maxReplicas)
			spec.Behavior.ScaleUp.SelectPolicy = v1alpha1.DisabledPolicySelect
			spec.Behavior.ScaleDown.SelectPolicy = v1alpha1.DisabledPolicySelect

			status := createStatus(currentReplicas, recommendations)

			normalizeAndVerifyResult(spec, status, desiredReplicas, expectedReplicas)
		},

		Entry("scale up disabled", 1, 5, 2, 5, []interface{}{}, 2),
		Entry("scale down disabled", 4, 10, 5, 2, []interface{}{1, 1, 2}, 5),
	)

	DescribeTable("ScaleUp disabled, ScaleDown fast",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := createSpec(minReplicas, maxReplicas)
			spec.Behavior.ScaleUp.SelectPolicy = v1alpha1.DisabledPolicySelect
			spec.Behavior.ScaleDown.SelectPolicy = v1alpha1.MaxPolicySelect
			spec.Behavior.ScaleDown.Policies = []v1alpha1.ScalingPolicy{
				createPolicy(v1alpha1.PodsScalingPolicy, 10, 15),
				createPolicy(v1alpha1.PercentScalingPolicy, 50, 15),
			}

			status := createStatus(currentReplicas, recommendations)

			normalizeAndVerifyResult(spec, status, desiredReplicas, expectedReplicas)
		},

		Entry("scale up disabled", 1, 10, 3, 5, []interface{}{}, 3),
		Entry("scale down 50% from current value", 5, 110, 100, 10, []interface{}{}, 50),
		Entry("scale down by 10 pods from current value", 1, 100, 15, 2, []interface{}{}, 5),
	)

	DescribeTable("ScaleUp disabled, ScaleDown slow",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := createSpec(minReplicas, maxReplicas)
			spec.Behavior.ScaleUp.SelectPolicy = v1alpha1.DisabledPolicySelect
			spec.Behavior.ScaleDown.SelectPolicy = v1alpha1.MinPolicySelect
			spec.Behavior.ScaleDown.Policies = []v1alpha1.ScalingPolicy{
				createPolicy(v1alpha1.PodsScalingPolicy, 5, 15),
				createPolicy(v1alpha1.PercentScalingPolicy, 50, 15),
			}

			status := createStatus(currentReplicas, recommendations)

			normalizeAndVerifyResult(spec, status, desiredReplicas, expectedReplicas)
		},

		Entry("scale up disabled", 1, 10, 3, 5, []interface{}{}, 3),
		Entry("scale down by 5 from current value", 5, 100, 100, 10, []interface{}{}, 95),
		Entry("scale down by 50% pods from current value", 1, 100, 6, 1, []interface{}{}, 3),
	)

	DescribeTable("ScaleUp fast, ScaleDown disabled",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := createSpec(minReplicas, maxReplicas)
			spec.Behavior.ScaleDown.SelectPolicy = v1alpha1.DisabledPolicySelect
			spec.Behavior.ScaleUp.SelectPolicy = v1alpha1.MaxPolicySelect
			spec.Behavior.ScaleUp.Policies = []v1alpha1.ScalingPolicy{
				createPolicy(v1alpha1.PodsScalingPolicy, 10, 15),
				createPolicy(v1alpha1.PercentScalingPolicy, 50, 15),
			}

			status := createStatus(currentReplicas, recommendations)

			normalizeAndVerifyResult(spec, status, desiredReplicas, expectedReplicas)
		},

		Entry("scale down disabled", 1, 10, 7, 4, []interface{}{}, 7),
		Entry("scale up 50% from current value", 5, 60, 30, 50, []interface{}{}, 45),
		Entry("scale up by 10 pods from current value", 1, 60, 15, 40, []interface{}{}, 25),
	)

	DescribeTable("ScaleUp slow, ScaleDown disabled",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := createSpec(minReplicas, maxReplicas)
			spec.Behavior.ScaleDown.SelectPolicy = v1alpha1.DisabledPolicySelect
			spec.Behavior.ScaleUp.SelectPolicy = v1alpha1.MinPolicySelect
			spec.Behavior.ScaleUp.Policies = []v1alpha1.ScalingPolicy{
				createPolicy(v1alpha1.PodsScalingPolicy, 10, 15),
				createPolicy(v1alpha1.PercentScalingPolicy, 50, 15),
			}

			status := createStatus(currentReplicas, recommendations)

			normalizeAndVerifyResult(spec, status, desiredReplicas, expectedReplicas)
		},

		Entry("scale down disabled", 1, 10, 7, 4, []interface{}{}, 7),
		Entry("scale up 50% from current value", 5, 60, 10, 50, []interface{}{}, 15),
		Entry("scale up by 10 pods from current value", 1, 60, 30, 50, []interface{}{}, 40),
	)

})

func createSpec(minReplicas int, maxReplicas int) *v1alpha1.KratosSpec {
	return &v1alpha1.KratosSpec{
		MinReplicas:                int32(minReplicas),
		MaxReplicas:                int32(maxReplicas),
		StabilizationWindowSeconds: 15,
		Behavior: &v1alpha1.ScaleBehavior{
			ScaleUp: &v1alpha1.ScaleRules{
				StabilizationWindowSeconds: 15,
			},

			ScaleDown: &v1alpha1.ScaleRules{
				StabilizationWindowSeconds: 15,
			},
		},
	}
}

func createStatus(currentReplicas int, recommendations []interface{}) *v1alpha1.KratosStatus {
	recordedRecommendations := make([]v1alpha1.Recommendation, len(recommendations))

	for i, item := range recommendations {
		recordedRecommendations[i] = v1alpha1.Recommendation{Timestamp: metav1.Now(), Replicas: int32(item.(int))}
	}

	return &v1alpha1.KratosStatus{
		CurrentReplicas: int32(currentReplicas),
		Recommendations: recordedRecommendations,
	}
}

func normalizeAndVerifyResult(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus, desiredReplicas int, expectedReplicas int) {
	normalizedReplicas := newBehaviorNormalizer().normalizeReplicas(spec, status, int32(desiredReplicas))
	Expect(normalizedReplicas).To(Equal(int32(expectedReplicas)))

	normalizedReplicasFacade := NewReplicaNormalizer().NormalizeReplicas(spec, status, int32(desiredReplicas))
	Expect(normalizedReplicasFacade).To(Equal(int32(expectedReplicas)))
}

func createPolicy(policyType v1alpha1.ScalingPolicyType, value int, windowSeconds int) v1alpha1.ScalingPolicy {
	return v1alpha1.ScalingPolicy{
		Type:          policyType,
		Value:         int32(value),
		PeriodSeconds: int32(windowSeconds),
	}
}
