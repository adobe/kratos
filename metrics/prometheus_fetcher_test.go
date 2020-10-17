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
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/adobe/kratos/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/model"
)

type queryResult struct {
	Type   model.ValueType `json:"resultType"`
	Result interface{}     `json:"result"`
}

type apiResponse struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType string          `json:"errorType"`
	Error     string          `json:"error"`
	Warnings  []string        `json:"warnings,omitempty"`
}

var _ = Describe("PrometheusFetcher", func() {
	var testServer *httptest.Server
	var fetcher MetricsFetcher
	var queryResults map[string]queryResult

	BeforeEach(func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.ParseForm()

			queryKey := req.Form.Get("query")
			results, _ := queryResults[queryKey]

			testResp, _ := json.Marshal(results)

			apiResp := &apiResponse{
				Data: testResp,
			}

			body, _ := json.Marshal(apiResp)

			w.Write(body)
		}))
		fetcher = newPrometheusMetricsFetcher(testServer.URL)
		queryResults = make(map[string]queryResult, 0)
	})

	AfterEach(func() {
		testServer.Close()
	})

	It("No prometheus response", func() {
		scaleMetric := &v1alpha1.ScaleMetric{
			Prometheus: &v1alpha1.PrometheusMetricSource{
				MetricQuery: "count(up)",
			}}
		_, err := fetcher.Fetch(scaleMetric)

		Expect(err).ToNot(BeNil(), "non prometheus response should result in error")
	})

	It("Scalar value", func() {
		query := "count(up)"
		queryResults[query] = queryResult{
			Type: model.ValScalar,
			Result: model.Scalar{
				Value:     55,
				Timestamp: model.Now(),
			}}
		scaleMetric := &v1alpha1.ScaleMetric{
			Prometheus: &v1alpha1.PrometheusMetricSource{
				MetricQuery: query,
			}}
		fetchResults, err := fetcher.Fetch(scaleMetric)

		Expect(err).To(BeNil(), "no errors on scalar value")
		Expect(len(fetchResults)).To(Equal(1), "scalar value should result in single item")
		Expect(fetchResults[0].Value).To(Equal(int64(55)), "metric value should match scalar value")
	})

	It("Vector value - empty", func() {
		query := "up"
		queryResults[query] = queryResult{
			Type:   model.ValVector,
			Result: make([]model.Sample, 0),
		}
		scaleMetric := &v1alpha1.ScaleMetric{
			Prometheus: &v1alpha1.PrometheusMetricSource{
				MetricQuery: query,
			}}
		fetchResults, err := fetcher.Fetch(scaleMetric)

		Expect(err).To(BeNil(), "no errors on vector value")
		Expect(len(fetchResults)).To(Equal(0), "empty vector value should result in empty metrics")
	})

	It("Vector value - not empty", func() {
		query := "up"
		samples := make([]model.Sample, 2)

		samples[0] = model.Sample{
			Value:     10,
			Timestamp: model.Now(),
		}

		samples[1] = model.Sample{
			Value:     20,
			Timestamp: model.Now(),
		}

		queryResults[query] = queryResult{
			Type:   model.ValVector,
			Result: samples,
		}

		scaleMetric := &v1alpha1.ScaleMetric{
			Prometheus: &v1alpha1.PrometheusMetricSource{
				MetricQuery: query,
			}}
		fetchResults, err := fetcher.Fetch(scaleMetric)

		Expect(err).To(BeNil(), "no errors on vector value")
		Expect(len(fetchResults)).To(Equal(len(samples)), "vector size should be equal to returned metrics size")
	})
})
