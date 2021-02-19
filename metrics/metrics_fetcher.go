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
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/adobe/kratos/api/common"
	"github.com/adobe/kratos/api/v1alpha1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	defaultCacheTtl    = time.Minute * 5
	defaultCallTimeout = time.Second * 10
)

type MetricValue struct {
	Value int64
}

type MetricsFetcher interface {
	Fetch(scaleMetric *v1alpha1.ScaleMetric, namespace string, selector labels.Selector) ([]MetricValue, error)
}

type MetricsFactory struct {
	prometheusFetcher MetricsFetcher
	resourceFetcher   MetricsFetcher
}

func NewMetricsFactory(params *common.KratosParameters) *MetricsFactory {
	mc, err := metrics.NewForConfig(params.ClientConfig)
	if err != nil {
		panic(err.Error())
	}
	return &MetricsFactory{
		prometheusFetcher: newPrometheusMetricsFetcher(params.DefaultPrometheusUrl),
		resourceFetcher:   newResourceMetricsFetcher(mc),
	}
}

func (facade *MetricsFactory) GetMetricsFetcher(scaleMetric *v1alpha1.ScaleMetric) (MetricsFetcher, error) {
	switch scaleMetric.Type {
	case v1alpha1.PrometheusScaleMetricType:
		return facade.prometheusFetcher, nil
	case v1alpha1.ResourceScaleMetricType:
		return facade.resourceFetcher, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown metric type %s \n", scaleMetric.Type))
	}
}
