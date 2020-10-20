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

import "github.com/adobe/kratos/api/v1alpha1"

type normalizer interface {
	normalizeReplicas(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus, desiredReplicas int32) int32
}

type ReplicaNormalizer struct {
	standardNormalizer  normalizer
	behaviourNormalizer normalizer
}

func NewReplicaNormalizer() *ReplicaNormalizer {
	return &ReplicaNormalizer{
		standardNormalizer:  newStandardNormalizer(),
		behaviourNormalizer: newBehaviorNormalizer(),
	}
}

func (n *ReplicaNormalizer) NormalizeReplicas(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus,
	desiredReplicas int32) int32 {
	if spec.Behavior == nil || (spec.Behavior.ScaleUp == nil && spec.Behavior.ScaleDown == nil) {
		return n.standardNormalizer.normalizeReplicas(spec, status, desiredReplicas)
	}

	return n.behaviourNormalizer.normalizeReplicas(spec, status, desiredReplicas)
}
