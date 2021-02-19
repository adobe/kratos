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
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/cache"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type prometheusMetricsFetcher struct {
	defaultUrl             string
	prometheusClientsCache *cache.TTLCache
	log                    logr.Logger
}

type prometheusClient struct {
	v1.API
}

func (c *prometheusClient) Close() error {
	return nil
}

func newPrometheusMetricsFetcher(defaultUrl string) *prometheusMetricsFetcher {
	fetcher := &prometheusMetricsFetcher{
		defaultUrl:             defaultUrl,
		prometheusClientsCache: cache.NewTTLCache("prometheus-clients", defaultCacheTtl),
		log:                    log.Log.WithName("prom-fetcher"),
	}

	return fetcher
}

func (p *prometheusMetricsFetcher) Fetch(scaleMetric *v1alpha1.ScaleMetric, namespace string, selector labels.Selector) ([]MetricValue, error) {
	client, err := p.getOrCreateClient(scaleMetric.Prometheus.PrometheusEndpoint)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultCallTimeout)
	defer cancel()

	p.log.V(1).Info("fetching metrics", "url", scaleMetric.Prometheus.PrometheusEndpoint, "query", scaleMetric.Prometheus.MetricQuery)

	result, warnings, err := client.Query(ctx, scaleMetric.Prometheus.MetricQuery, time.Now())

	if err != nil {
		return nil, err
	}

	if warnings != nil {
		p.log.V(1).Info("warnings on fetching metrics", "url", scaleMetric.Prometheus.PrometheusEndpoint, "warnings", warnings)
	}
	return p.convertToMetricValue(result)
}

func (p *prometheusMetricsFetcher) getOrCreateClient(queryUrl string) (*prometheusClient, error) {
	prometheusUrl := p.defaultUrl

	if queryUrl != "" {
		prometheusUrl = queryUrl
	}

	cachedClient, found := p.prometheusClientsCache.Get(prometheusUrl)

	if !found {
		client, err := p.createPrometheusApi(prometheusUrl)

		if err != nil {
			return nil, err
		}

		p.prometheusClientsCache.Put(queryUrl, client)
		cachedClient = client
	}

	castedClient := cachedClient.(*prometheusClient)
	return castedClient, nil
}

func (p *prometheusMetricsFetcher) createPrometheusApi(prometheusUrl string) (*prometheusClient, error) {
	client, err := api.NewClient(api.Config{
		Address: prometheusUrl,
	})

	if err != nil {
		p.log.Error(err, "can't create client for prometheus", "url", prometheusUrl)
		return nil, err
	}

	prometheusClient := &prometheusClient{v1.NewAPI(client)}
	return prometheusClient, nil
}

func (p *prometheusMetricsFetcher) convertToMetricValue(result model.Value) ([]MetricValue, error) {
	switch result.Type() {
	case model.ValScalar:
		return p.convertScalar(result)
	case model.ValVector:
		return p.convertVector(result)
	default:
		return nil, errors.New(fmt.Sprintf("unsupported prometheus result type: %v\n", result.Type()))
	}
}

func (p *prometheusMetricsFetcher) convertScalar(value model.Value) ([]MetricValue, error) {
	scalar, ok := value.(*model.Scalar)

	if !ok {
		return nil, errors.New("can't convert prometheus value to Scalar")
	}

	metricValue, err := p.convertSampleValue(scalar.Value)

	if err != nil {
		return nil, err
	}

	return []MetricValue{metricValue}, nil
}

func (p *prometheusMetricsFetcher) convertVector(value model.Value) ([]MetricValue, error) {
	vector, ok := value.(model.Vector)

	if !ok {
		return nil, errors.New("can't convert prometheus value to Vector")
	}

	result := make([]MetricValue, vector.Len())

	for i, sample := range vector {
		metricValue, err := p.convertSampleValue(sample.Value)
		if err != nil {
			return nil, err
		}
		result[i] = metricValue
	}

	return result, nil
}

func (p *prometheusMetricsFetcher) convertSampleValue(sample model.SampleValue) (MetricValue, error) {
	integerValue, err := strconv.ParseInt(sample.String(), 10, 64)

	if err != nil {
		return MetricValue{}, err
	}

	metricValue := MetricValue{
		Value: integerValue,
	}

	return metricValue, nil
}
