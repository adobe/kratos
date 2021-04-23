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
	"context"
	"fmt"

	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/client-go/util/retry"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type resourceMetricsFetcher struct {
	metricsClient clientset.Interface
	log           logr.Logger
}

func newResourceMetricsFetcher(metricsClient clientset.Interface) *resourceMetricsFetcher {
	fetcher := &resourceMetricsFetcher{
		metricsClient: metricsClient,
		log:           log.Log.WithName("resourceMetrics-fetcher"),
	}

	return fetcher
}

func (r *resourceMetricsFetcher) Fetch(scaleMetric *v1alpha1.ScaleMetric, namespace string, selector labels.Selector) ([]MetricValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCallTimeout)
	defer cancel()

	opts := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	podMetricsList, err := r.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	r.log.Info("Fetched resource metrics", "selector", selector.String(), "container", scaleMetric.Resource.Container, "metrics", podMetricsList.Items)

	ret := []MetricValue{}

	for _, podMetrics := range podMetricsList.Items {
		var accum int64 = 0
		found := false
		for _, container := range podMetrics.Containers {
			if scaleMetric.Resource.Container == "" || scaleMetric.Resource.Container == container.Name {
				found = true
				switch scaleMetric.Resource.Name {
				case corev1.ResourceCPU:
					accum += container.Usage.Cpu().Value()
				case corev1.ResourceMemory:
					accum += container.Usage.Memory().Value()
				default:
					return nil, fmt.Errorf("Unsuported resource type %s", scaleMetric.Resource.Name)
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("container %s not present in metrics for pod %s/%s", scaleMetric.Resource.Container, namespace, podMetrics.Name)
		}
		ret = append(ret, MetricValue{accum})
	}

	return ret, nil
}
