/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package scale

import (
	"github.com/adobe/kratos/pkg/api/common"
	"github.com/adobe/kratos/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultsUpdater", func() {
	It("Stabilization window", func() {
		params := &common.KratosParameters{
			StabilizationWindowSeconds: 200,
		}

		updater := newDefaultsUpdater(params)

		spec := &v1alpha1.KratosSpec{}

		Expect(spec.StabilizationWindowSeconds).To(Equal(int32(0)))
		updater.updateSpecWithDefaults(spec)

		Expect(spec.StabilizationWindowSeconds).To(Equal(params.StabilizationWindowSeconds), "stabilization should be set to default")
		Expect(spec.Behavior).To(BeNil())
	})

	It("Negative MinReplicas", func() {
		params := &common.KratosParameters{
			StabilizationWindowSeconds: 200,
		}

		updater := newDefaultsUpdater(params)

		spec := &v1alpha1.KratosSpec{
			MinReplicas: -10,
		}

		Expect(spec.MinReplicas).To(Equal(int32(-10)))
		updater.updateSpecWithDefaults(spec)
		Expect(spec.MinReplicas).To(Equal(int32(0)), "negative min replicas should be set to 0")
	})

	It("Negative MaxReplicas", func() {
		params := &common.KratosParameters{
			StabilizationWindowSeconds: 200,
		}

		updater := newDefaultsUpdater(params)

		spec := &v1alpha1.KratosSpec{
			MaxReplicas: -10,
		}

		Expect(spec.MaxReplicas).To(Equal(int32(-10)))
		updater.updateSpecWithDefaults(spec)
		Expect(spec.MaxReplicas).To(Equal(int32(0)), "negative max replicas should be set to 0")
	})

	It("Scale policies", func() {
		params := &common.KratosParameters{
			StabilizationWindowSeconds: 200,
		}

		updater := newDefaultsUpdater(params)

		behavior := &v1alpha1.ScaleBehavior{}
		behavior.ScaleUp = &v1alpha1.ScaleRules{
			StabilizationWindowSeconds: -1,
			SelectPolicy:               v1alpha1.DisabledPolicySelect,
		}

		behavior.ScaleDown = &v1alpha1.ScaleRules{
			StabilizationWindowSeconds: -15,
		}

		spec := &v1alpha1.KratosSpec{
			Behavior: behavior,
		}

		updater.updateSpecWithDefaults(spec)
		Expect(behavior.ScaleUp.StabilizationWindowSeconds).To(Equal(int32(-1)), "stabilization window should not be updated on  disabled scale")
		Expect(behavior.ScaleDown.StabilizationWindowSeconds).To(Equal(params.StabilizationWindowSeconds), "stabilization window should be updated to default")
		Expect(behavior.ScaleDown.SelectPolicy).To(Equal(v1alpha1.MinPolicySelect), "default scale down select policy should be MinPolicySelect")
	})
})
