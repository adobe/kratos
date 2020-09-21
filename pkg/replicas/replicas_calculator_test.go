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
	"github.com/adobe/kratos/pkg/api/v1alpha1"
	"github.com/adobe/kratos/pkg/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("ReplicaCalculator", func() {
	It("Not implemented target metric type", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		metricTarget := v1alpha1.MetricTarget{
			Type: v1alpha1.UtilizationMetricType,
		}

		_, err := replicaCalculator.CalculateReplicas(1, metricTarget, nil)

		Expect(err).ToNot(BeNil(), "not implemented metric target should result in error")
	})

	It("Target value calculator", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		quantity, _ := resource.ParseQuantity("5")
		metricTarget := v1alpha1.MetricTarget{
			Type:  v1alpha1.ValueMetricType,
			Value: &quantity,
		}

		metrics := []metrics.MetricValue{
			metrics.MetricValue{
				Value: 10,
			},
		}
		replicas, err := replicaCalculator.CalculateReplicas(1, metricTarget, metrics)

		Expect(err).To(BeNil(), "no errors expected for valid arguments")
		Expect(replicas).To(Equal(int32(2)), "wrong NR. of replicas")
	})

})
