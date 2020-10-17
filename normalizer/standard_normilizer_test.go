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

var _ = Describe("StandardNormalizer", func() {
	var normalizer = newStandardNormalizer()
	var normalizerFacade = NewReplicaNormalizer()

	DescribeTable("Standard normalizer",
		func(minReplicas int, maxReplicas int, currentReplicas int, desiredReplicas int, recommendations []interface{}, expectedReplicas int) {
			spec := &v1alpha1.KratosSpec{
				StabilizationWindowSeconds: 15,
				MinReplicas:                int32(minReplicas),
				MaxReplicas:                int32(maxReplicas),
			}

			recordedRecommendations := make([]v1alpha1.Recommendation, len(recommendations))

			for i, item := range recommendations {
				recordedRecommendations[i] = v1alpha1.Recommendation{Timestamp: metav1.Now(), Replicas: int32(item.(int))}
			}

			status := &v1alpha1.KratosStatus{
				CurrentReplicas: int32(currentReplicas),
				Recommendations: recordedRecommendations,
			}

			normalizedReplicas := normalizer.normalizeReplicas(spec, status, int32(desiredReplicas))
			Expect(normalizedReplicas).To(Equal(int32(expectedReplicas)))

			normalizedReplicasFacade := normalizerFacade.NormalizeReplicas(spec, status, int32(desiredReplicas))
			Expect(normalizedReplicasFacade).To(Equal(int32(expectedReplicas)))
		},

		Entry("scale factor produces max replicas ", 1, 5, 2, 5, []interface{}{}, 4),
		Entry("scale limited by max replicas ", 1, 10, 9, 11, []interface{}{4, 15}, 10),
		Entry("scale limited by min replicas ", 4, 10, 5, 2, []interface{}{1, 1, 2}, 4),
	)
})
