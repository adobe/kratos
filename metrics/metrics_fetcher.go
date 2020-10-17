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

	"github.com/adobe/kratos/api/v1alpha1"
)

type MetricValue struct {
	Value int64
}

type MetricsFetcher interface {
	Fetch(scaleMetric *v1alpha1.ScaleMetric) ([]MetricValue, error)
}

type MetricsFactory struct {
	prometheusFetcher MetricsFetcher
}

func NewMetricsFactory(defaultPrometheusUrl string) *MetricsFactory {
	return &MetricsFactory{
		prometheusFetcher: newPrometheusMetricsFetcher(defaultPrometheusUrl),
	}
}

func (facade *MetricsFactory) GetMetricsFetcher(scaleMetric *v1alpha1.ScaleMetric) (MetricsFetcher, error) {
	switch scaleMetric.Type {
	case v1alpha1.PrometheusScaleMetricType:
		return facade.prometheusFetcher, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown metric type %s \n", scaleMetric.Type))
	}
}
