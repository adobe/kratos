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
	"github.com/adobe/kratos/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MetricsFactory", func() {

	It("Unknown fetcher type", func() {

		metricsFactory := NewMetricsFactory(fakeKratosSpec)
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: "testType",
		}
		_, err := metricsFactory.GetMetricsFetcher(scaleMetric)

		Expect(err).NotTo(BeNil(), "unknown metric type should result in error")
	})

	It("Prometheus fetcher type", func() {
		metricsFactory := NewMetricsFactory(fakeKratosSpec)
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.PrometheusScaleMetricType,
		}
		fetcher, err := metricsFactory.GetMetricsFetcher(scaleMetric)

		Expect(err).To(BeNil(), "no error for supported metrics fetcher type")
		Expect(fetcher).NotTo(BeNil())
	})

	It("Resource fetcher type", func() {
		metricsFactory := NewMetricsFactory(fakeKratosSpec)
		scaleMetric := &v1alpha1.ScaleMetric{
			Type: v1alpha1.ResourceScaleMetricType,
		}
		fetcher, err := metricsFactory.GetMetricsFetcher(scaleMetric)

		Expect(err).To(BeNil(), "no error for supported metrics fetcher type")
		Expect(fetcher).NotTo(BeNil())
	})
})
