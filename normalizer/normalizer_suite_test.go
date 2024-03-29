/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package normalizer

import (
	"testing"

	"github.com/adobe/kratos/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicas(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Normalizer Suite")
}

var _ = BeforeSuite(func() {

}, 60)

var _ = AfterSuite(func() {

})

func appendRecommendations(status *v1alpha1.KratosStatus, recommendations ...int32) {
	status.Recommendations = make([]v1alpha1.Recommendation, len(recommendations))
	for i, value := range recommendations {
		status.Recommendations[i] = v1alpha1.Recommendation{
			Timestamp: metav1.Now(),
			Replicas:  value,
		}
	}
}
