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
	"context"
	"fmt"

	"github.com/adobe/kratos/pkg/api/common"
	"github.com/adobe/kratos/pkg/api/v1alpha1"
	"github.com/go-logr/logr"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScaleTarget struct {
	client       client.Client
	log          logr.Logger
	mapper       meta.RESTMapper
	scalesGetter scale.ScalesGetter
}

func NewScaleTarget(params *common.KratosParameters) (*ScaleTarget, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(params.ClientConfig)
	if err != nil {
		return nil, err
	}

	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	scalesGetter, err := scale.NewForConfig(params.ClientConfig, params.RestMapper, dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		return nil, err
	}

	facade := &ScaleTarget{
		client:       params.Client,
		log:          ctrl.Log.WithName("target-facade"),
		mapper:       params.RestMapper,
		scalesGetter: scalesGetter,
	}

	return facade, nil
}

func (st *ScaleTarget) GetScaleTarget(namespace string, targetRef *v1alpha1.ScaleTargetReference) (*autoscalingv1.Scale, *schema.GroupResource, error) {
	reference := fmt.Sprintf("%s/%s/%s", targetRef.Kind, namespace, targetRef.Name)

	targetGroupVersion, err := schema.ParseGroupVersion(targetRef.APIVersion)

	if err != nil {
		return nil, &schema.GroupResource{}, fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGroupKind := schema.GroupKind{
		Group: targetGroupVersion.Group,
		Kind:  targetRef.Kind,
	}

	mappings, err := st.mapper.RESTMappings(targetGroupKind)

	if err != nil {
		return nil, &schema.GroupResource{}, fmt.Errorf("unable to determine resource for scale target reference: %v", err)
	}

	scaleObject, targetGroupResource, err := st.scaleForResourceMappings(namespace, targetRef.Name, mappings)

	if err != nil {
		return nil, &schema.GroupResource{}, fmt.Errorf("failed to query scale subresource for %s: %v", reference, err)
	}

	return scaleObject, &targetGroupResource, nil
}

func (st *ScaleTarget) Scale(namespace string, groupResource *schema.GroupResource, scale *autoscalingv1.Scale) error {
	st.log.V(1).Info("scaling target", "namespace", namespace, "name", scale.Name)
	_, err := st.scalesGetter.Scales(namespace).Update(context.TODO(), *groupResource, scale, metav1.UpdateOptions{})
	return err
}

func (st *ScaleTarget) scaleForResourceMappings(namespace, name string, mappings []*meta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scaleObject, err := st.scalesGetter.Scales(namespace).Get(context.TODO(), targetGR, name, metav1.GetOptions{})
		if err == nil {
			return scaleObject, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
}
