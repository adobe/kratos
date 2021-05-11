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
	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/metrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var noRequestedResources = map[string]*corev1.ResourceList{
	"": {
		corev1.ResourceCPU:     *resource.NewQuantity(0, resource.DecimalSI),
		corev1.ResourceMemory:  *resource.NewQuantity(0, resource.BinarySI),
		corev1.ResourceStorage: *resource.NewQuantity(0, resource.BinarySI),
	},
	"nginx": {
		corev1.ResourceCPU:     *resource.NewQuantity(0, resource.DecimalSI),
		corev1.ResourceMemory:  *resource.NewQuantity(0, resource.BinarySI),
		corev1.ResourceStorage: *resource.NewQuantity(0, resource.BinarySI),
	},
}
var requestedResources = map[string]*corev1.ResourceList{
	"": {
		corev1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
		corev1.ResourceMemory:  *resource.NewQuantity(1000, resource.BinarySI),
		corev1.ResourceStorage: *resource.NewQuantity(10, resource.BinarySI),
	},
	"nginx": {
		corev1.ResourceCPU:     *resource.NewQuantity(10, resource.DecimalSI),
		corev1.ResourceMemory:  *resource.NewQuantity(1000, resource.BinarySI),
		corev1.ResourceStorage: *resource.NewQuantity(10, resource.BinarySI),
	},
}

var _ = Describe("ReplicaCalculator", func() {
	It("Not implemented target metric type", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		scaleMetric := v1alpha1.ScaleMetric{
			Type: v1alpha1.ExternalScaleMetricType,
			Prometheus: &v1alpha1.PrometheusMetricSource{
				Target: v1alpha1.MetricTarget{
					Type:         v1alpha1.AverageValueMetricType,
					AverageValue: resource.NewMilliQuantity(500, resource.DecimalSI),
				},
			},
		}

		_, err := replicaCalculator.CalculateReplicas(1, noRequestedResources, scaleMetric, nil)

		Expect(err).ToNot(BeNil(), "not implemented metric target should result in error")
	})

	It("Prometheus target value calculator", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		scaleMetric := v1alpha1.ScaleMetric{
			Type: v1alpha1.PrometheusScaleMetricType,
			Prometheus: &v1alpha1.PrometheusMetricSource{
				Target: v1alpha1.MetricTarget{
					Type:         v1alpha1.AverageValueMetricType,
					AverageValue: resource.NewQuantity(5, resource.DecimalSI),
				},
			},
		}

		metrics := []metrics.MetricValue{
			{
				Value: 10,
			},
		}
		replicas, err := replicaCalculator.CalculateReplicas(1, noRequestedResources, scaleMetric, metrics)

		Expect(err).To(BeNil(), "no errors expected for valid arguments")
		Expect(replicas).To(Equal(int32(2)), "wrong number of replicas")
	})

	It("Resource target - no requests specified", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		averageUtilization := int32(50)
		scaleMetric := v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: v1alpha1.MetricTarget{
					Type:               v1alpha1.UtilizationMetricType,
					AverageUtilization: &averageUtilization,
				},
			},
		}

		metrics := []metrics.MetricValue{
			{
				Value: 10,
			},
		}
		_, err := replicaCalculator.CalculateReplicas(1, noRequestedResources, scaleMetric, metrics)

		Expect(err).ToNot(BeNil(), "should fail proposal for pods without requested resources")
	})

	It("Resource target - value calculator", func() {
		replicaCalculator := NewReplicaCalculator(0.1)

		averageUtilization := int32(50)
		scaleMetric := v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
				Target: v1alpha1.MetricTarget{
					Type:               v1alpha1.UtilizationMetricType,
					AverageUtilization: &averageUtilization,
				},
			},
		}

		metrics := []metrics.MetricValue{
			{
				Value: 10,
			},
		}
		replicas, err := replicaCalculator.CalculateReplicas(1, requestedResources, scaleMetric, metrics)

		Expect(err).To(BeNil(), "no errors expected for valid arguments")
		Expect(replicas).To(Equal(int32(2)), "wrong number of replicas")
	})
})
