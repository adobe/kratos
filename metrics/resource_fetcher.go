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

	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/cache"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "k8s.io/client-go/util/retry"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type resourceMetricsFetcher struct {
	client               clientset.Interface
	resourceMetricsCache *cache.TTLCache
	log                  logr.Logger
}

func newResourceMetricsFetcher(client clientset.Interface) *resourceMetricsFetcher {
	fetcher := &resourceMetricsFetcher{
		client:               client,
		resourceMetricsCache: cache.NewTTLCache("resourceMetrics-clients", defaultCacheTtl),
		log:                  log.Log.WithName("resourceMetrics-fetcher"),
	}

	return fetcher
}

func (p *resourceMetricsFetcher) Fetch(scaleMetric *v1alpha1.ScaleMetric, namespace string, selector labels.Selector) ([]MetricValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCallTimeout)
	defer cancel()

	opt := metav1.ListOptions{
		LabelSelector: selector.String(),
	}

	podList, err := p.client.MetricsV1beta1().PodMetricses(namespace).List(ctx, opt)
	if err != nil {
		return nil, err
	}
	ret := []MetricValue{}

	for _, podMetrics := range podList.Items {
		for _, container := range podMetrics.Containers {
			switch scaleMetric.Resource.Name {
			case corev1.ResourceCPU:
				ret = append(ret, MetricValue{container.Usage.Cpu().Value()})
			case corev1.ResourceMemory:
				ret = append(ret, MetricValue{container.Usage.Memory().Value()})
			}
		}
	}

	return ret, nil
}
