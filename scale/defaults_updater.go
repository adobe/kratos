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
	"github.com/adobe/kratos/api/common"
	"github.com/adobe/kratos/api/v1alpha1"
)

type DefaultsUpdater struct {
	stabilizationWindowSeconds int32
}

func newDefaultsUpdater(params *common.KratosParameters) *DefaultsUpdater {
	return &DefaultsUpdater{
		stabilizationWindowSeconds: params.StabilizationWindowSeconds,
	}
}

func (p *DefaultsUpdater) updateSpecWithDefaults(spec *v1alpha1.KratosSpec) {
	p.updateStabilizationWindow(spec)
	p.updateReplicas(spec)
	p.updateScaleRules(spec)
}

func (p *DefaultsUpdater) updateReplicas(spec *v1alpha1.KratosSpec) {
	if spec.MinReplicas < 0 {
		spec.MinReplicas = 0
	}

	if spec.MaxReplicas < 0 {
		spec.MaxReplicas = 0
	}
}

func (p *DefaultsUpdater) updateStabilizationWindow(spec *v1alpha1.KratosSpec) {
	if spec.StabilizationWindowSeconds <= 0 {
		spec.StabilizationWindowSeconds = p.stabilizationWindowSeconds
	}
}

func (p *DefaultsUpdater) updateScaleRules(spec *v1alpha1.KratosSpec) {
	if spec.Behavior == nil {
		return
	}

	if spec.Behavior.ScaleUp != nil {
		scaleRules := spec.Behavior.ScaleUp
		if scaleRules.SelectPolicy == "" {
			scaleRules.SelectPolicy = v1alpha1.MaxPolicySelect
		}

		if scaleRules.StabilizationWindowSeconds <= 0 && scaleRules.SelectPolicy != v1alpha1.DisabledPolicySelect {
			scaleRules.StabilizationWindowSeconds = p.stabilizationWindowSeconds
		}
	} else {
		spec.Behavior.ScaleUp = &v1alpha1.ScaleRules{
			SelectPolicy: v1alpha1.DisabledPolicySelect,
		}
	}

	if spec.Behavior.ScaleDown != nil {
		scaleRules := spec.Behavior.ScaleDown
		if scaleRules.SelectPolicy == "" {
			scaleRules.SelectPolicy = v1alpha1.MinPolicySelect
		}

		if scaleRules.StabilizationWindowSeconds <= 0 && scaleRules.SelectPolicy != v1alpha1.DisabledPolicySelect {
			scaleRules.StabilizationWindowSeconds = p.stabilizationWindowSeconds
		}
	} else {
		spec.Behavior.ScaleDown = &v1alpha1.ScaleRules{
			SelectPolicy: v1alpha1.DisabledPolicySelect,
		}
	}
}
