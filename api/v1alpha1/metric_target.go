/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package v1alpha1

import (
	"fmt"
)

func (sm *ScaleMetric) GetMetricTarget() (*MetricTarget, error) {
	switch sm.Type {
	case PrometheusScaleMetricType:
		return &sm.Prometheus.Target, nil
	case ResourceScaleMetricType:
		return &sm.Resource.Target, nil
	default:
		return nil, fmt.Errorf("unknown metric type %s", sm.Type)
	}
}
