/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package metrics

import (
	"fmt"
	"time"

	"github.com/adobe/kratos/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var podNamePrefix string = "nginx"
var namespace string = "default"
var numFakePods int = 5

func makePodMetric(id int, labelSet map[string]string) metricsapi.PodMetrics {
	return metricsapi.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", podNamePrefix, id),
			Namespace: namespace,
			Labels:    labelSet,
		},
		Timestamp: metav1.Time{Time: time.Now()},
		Window:    metav1.Duration{Duration: time.Minute},
		Containers: []metricsapi.ContainerMetrics{
			{
				Name: podNamePrefix,
				Usage: v1.ResourceList{
					v1.ResourceCPU: *resource.NewQuantity(
						int64(id),
						resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(
						int64(id*1024*1024),
						resource.BinarySI),
				},
			},
		},
	}
}

var _ = Describe("ResourceFetcher", func() {
	fakeMetricsClient := &fake.Clientset{}

	fakeMetricsClient.AddReactor("list", "pods", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
		// tc.Lock()
		// defer tc.Unlock()

		metrics := &metricsapi.PodMetricsList{}

		for i := 1; i <= numFakePods; i++ {
			podMetric := makePodMetric(i, map[string]string{"app": "nginx"})
			metrics.Items = append(metrics.Items, podMetric)
			podMetric = makePodMetric(i+10, map[string]string{"app": "nginx", "state": "backup"})
			metrics.Items = append(metrics.Items, podMetric)
		}

		return true, metrics, nil
	})

	fetcher := newResourceMetricsFetcher(fakeMetricsClient)

	It("No resource response", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
			},
		}
		selector := labels.SelectorFromSet(labels.Set{
			"app": "nonexistent",
		})
		res, err := fetcher.Fetch(scaleMetric, namespace, selector)

		Expect(res, err).To(BeEmpty(), "get no metrics with Nothing label selector")
	})

	It("All resources response", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
			},
		}
		res, err := fetcher.Fetch(scaleMetric, namespace, labels.Everything())

		Expect(res, err).ToNot(BeEmpty(), "get all namespaced metrics should return result")
	})

	It("CPU resource", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
			},
		}
		res, err := fetcher.Fetch(scaleMetric, namespace, labels.Everything())

		Expect(res, err).ToNot(BeEmpty(), "get all namespaced metrics should return result")
	})

	It("Memory resource", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceMemory,
			}}
		res, err := fetcher.Fetch(scaleMetric, namespace, labels.Everything())

		Expect(res, err).ToNot(BeEmpty(), "get all namespaced metrics should return result")
	})

	It("Resource Metrics with selector", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
			Resource: &v1alpha1.ResourceMetricSource{
				Name: corev1.ResourceCPU,
			}}
		selector := labels.SelectorFromSet(labels.Set{
			"state": "backup",
		})
		res, err := fetcher.Fetch(scaleMetric, namespace, selector)

		Expect(res, err).ToNot(BeEmpty(), "get metrics with selector should return result")
		Expect(len(res)).To(Equal(int(numFakePods)), "get the expected number of metrics")
	})
})
