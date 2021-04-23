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
	"time"

	"github.com/adobe/kratos/api/common"
	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScaleTarget struct {
	client       client.Client
	log          logr.Logger
	mapper       meta.RESTMapper
	scalesGetter scale.ScalesGetter
	informersMap map[string]cache.SharedIndexInformer
}

const (
	daemonSet             string = "DaemonSet"
	deployment            string = "Deployment"
	replicaSet            string = "ReplicaSet"
	statefulSet           string = "StatefulSet"
	replicationController string = "ReplicationController"
	job                   string = "Job"
	cronJob               string = "CronJob"
)

const defaultResyncPeriod time.Duration = 10 * time.Minute

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

	log := ctrl.Log.WithName("target-facade")

	clientset, err := kubernetes.NewForConfig(params.ClientConfig)
	if err != nil {
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, defaultResyncPeriod)

	informersMap := map[string]cache.SharedIndexInformer{
		daemonSet:             factory.Apps().V1().DaemonSets().Informer(),
		deployment:            factory.Apps().V1().Deployments().Informer(),
		replicaSet:            factory.Apps().V1().ReplicaSets().Informer(),
		statefulSet:           factory.Apps().V1().StatefulSets().Informer(),
		replicationController: factory.Core().V1().ReplicationControllers().Informer(),
		job:                   factory.Batch().V1().Jobs().Informer(),
		cronJob:               factory.Batch().V1beta1().CronJobs().Informer(),
	}

	for kind, informer := range informersMap {
		stopCh := make(chan struct{})
		go informer.Run(stopCh)
		synced := cache.WaitForCacheSync(stopCh, informer.HasSynced)
		if !synced {
			log.Info(fmt.Sprintf("Could not sync cache for %s", kind))
		} else {
			log.Info(fmt.Sprintf("Initial sync of %s completed", kind))
		}
	}

	facade := &ScaleTarget{
		client:       params.Client,
		log:          log,
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

func (st *ScaleTarget) GetSelectorForTarget(namespace string, targetRef *v1alpha1.ScaleTargetReference) (labels.Selector, error) {
	informer, exists := st.informersMap[targetRef.Kind]
	if exists {
		return getLabelSelector(informer, targetRef.Kind, namespace, targetRef.Name)
	}

	targetGroupVersion, err := schema.ParseGroupVersion(targetRef.APIVersion)

	if err != nil {
		return nil, fmt.Errorf("invalid API version in scale target reference: %v", err)
	}

	targetGroupKind := schema.GroupKind{
		Group: targetGroupVersion.Group,
		Kind:  targetRef.Kind,
	}

	mappings, err := st.mapper.RESTMappings(targetGroupKind)

	selector, err := st.getLabelSelectorFromMappings(namespace, targetRef.Name, mappings)
	if err != nil {
		return nil, fmt.Errorf("Unhandled targetRef %v, error %v", targetRef, err)
	}

	return selector, nil
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

func (st *ScaleTarget) getLabelSelectorFromMappings(namespace, name string, mappings []*meta.RESTMapping) (labels.Selector, error) {
	var firstErr error
	for i, mapping := range mappings {
		groupResource := mapping.Resource.GroupResource()
		scale, err := st.scalesGetter.Scales(namespace).Get(context.TODO(), groupResource, name, metav1.GetOptions{})
		if err == nil {
			if scale.Status.Selector == "" {
				return nil, fmt.Errorf("Resource %s/%s has an empty selector for scale sub-resource", namespace, name)
			}
			selector, err := labels.Parse(scale.Status.Selector)
			if err != nil {
				return nil, err
			}
			return selector, nil
		}
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, firstErr
}

func getLabelSelector(informer cache.SharedIndexInformer, kind, namespace, name string) (labels.Selector, error) {
	obj, exists, err := informer.GetStore().GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("%s %s/%s does not exist", kind, namespace, name)
	}
	switch obj.(type) {
	case (*appsv1.DaemonSet):
		apiObj, ok := obj.(*appsv1.DaemonSet)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.Deployment):
		apiObj, ok := obj.(*appsv1.Deployment)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.StatefulSet):
		apiObj, ok := obj.(*appsv1.StatefulSet)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*appsv1.ReplicaSet):
		apiObj, ok := obj.(*appsv1.ReplicaSet)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*batchv1.Job):
		apiObj, ok := obj.(*batchv1.Job)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(apiObj.Spec.Selector)
	case (*batchv1beta1.CronJob):
		apiObj, ok := obj.(*batchv1beta1.CronJob)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.JobTemplate.Spec.Template.Labels))
	case (*corev1.ReplicationController):
		apiObj, ok := obj.(*corev1.ReplicationController)
		if !ok {
			return nil, fmt.Errorf("Failed to parse %s %s/%s", kind, namespace, name)
		}
		return metav1.LabelSelectorAsSelector(metav1.SetAsLabelSelector(apiObj.Spec.Selector))
	}
	return nil, fmt.Errorf("Don't know how to read label seletor")
}
