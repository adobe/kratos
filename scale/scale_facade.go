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
	"errors"
	"time"

	"github.com/adobe/kratos/api/common"
	"github.com/adobe/kratos/api/v1alpha1"
	"github.com/adobe/kratos/metrics"
	"github.com/adobe/kratos/normalizer"
	"github.com/adobe/kratos/replicas"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	empty     = ""
	specKey   = "kratosSpec"
	statusKey = "kratosStatus"
)

type ScaleFacade struct {
	client            client.Client
	log               logr.Logger
	scaleTarget       *ScaleTarget
	metricsFactory    *metrics.MetricsFactory
	replicaCalculator *replicas.ReplicaCalculator
	replicaNormalizer *normalizer.ReplicaNormalizer
	eventRecorder     record.EventRecorder
	defaultsUpdater   *DefaultsUpdater
}

func NewScaleFacade(params *common.KratosParameters) (*ScaleFacade, error) {
	scaleTarget, err := NewScaleTarget(params)

	if err != nil {
		return nil, err
	}

	facade := &ScaleFacade{
		client:            params.Client,
		log:               ctrl.Log.WithName("scale-facade"),
		scaleTarget:       scaleTarget,
		metricsFactory:    metrics.NewMetricsFactory(params.DefaultPrometheusUrl),
		replicaCalculator: replicas.NewReplicaCalculator(0.1), //todo params?
		replicaNormalizer: normalizer.NewReplicaNormalizer(),
		eventRecorder:     params.EventRecorder,
		defaultsUpdater:   newDefaultsUpdater(params),
	}

	return facade, nil
}

func (f *ScaleFacade) Scale(item *corev1.ConfigMap) {
	log := f.log.WithValues("namespace", item.GetNamespace(), "name", item.GetName())

	log.V(1).Info("unmarshalling")
	spec, status, err := f.unmarshall(item.Data)
	if err != nil {
		log.Error(err, "error on unmarshalling")
		f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "UnmarshallError", "can't unmarshall: %v", err.Error())
		return
	}

	log.V(1).Info("updating defaults")
	f.defaultsUpdater.updateSpecWithDefaults(spec)

	log.V(1).Info("retrieving scale target")
	scaleObject, groupResource, err := f.scaleTarget.GetScaleTarget(item.Namespace, &spec.Target)
	if err != nil {
		log.Error(err, "error on retrieving scale target")
		f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "RetrieveScaleTargetError", "can't retrieve scale target: %v", err.Error())
		return
	}

	currentReplicas := scaleObject.Status.Replicas
	status.CurrentReplicas = currentReplicas

	log.V(1).Info("retrieved scale target", "replicas", currentReplicas)

	defer func() {
		err := f.updateObjects(item, spec, status)
		if err != nil {
			log.Error(err, "Error on updating object")
		}
	}()

	f.expireRecommendationsAndScaleEvents(spec, status)

	log.V(1).Info("calculating max replicas using metrics")
	desiredReplicas := f.calculateMaxScaleReplicas(item, currentReplicas, spec)
	log.V(1).Info("desired max replicas", "replicas", desiredReplicas)

	log.V(1).Info("recording max replicas recommendation")
	f.recordRecommendation(desiredReplicas, status)

	log.V(1).Info("normalizing max replicas using behaviour policies")
	normalizedReplicas := f.replicaNormalizer.NormalizeReplicas(spec, status, desiredReplicas)
	log.V(1).Info("normalized replicas", "replicas", normalizedReplicas)
	f.eventRecorder.Eventf(item, corev1.EventTypeNormal, "CalculateReplicas", "replicas - current: %d, metrics: %d, normalized: %d", status.CurrentReplicas, desiredReplicas, normalizedReplicas)

	if normalizedReplicas != status.CurrentReplicas {
		scaleObject.Spec.Replicas = normalizedReplicas

		log.V(1).Info("scaling target", "namespace", scaleObject.GetNamespace(), "name", scaleObject.GetName(), "replicas", normalizedReplicas)
		err := f.scaleTarget.Scale(item.Namespace, groupResource, scaleObject)

		if err == nil {
			f.recordScaleEvent(currentReplicas, normalizedReplicas, status)
		} else {
			log.Error(err, "unable to scale target to desired replicas", "namespace", scaleObject.GetNamespace(), "name", scaleObject.GetName(), "replicas", normalizedReplicas)
			f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "ScaleError", "can't scale target: %v", err.Error())
		}

	}
}

func (f *ScaleFacade) unmarshall(data map[string]string) (*v1alpha1.KratosSpec, *v1alpha1.KratosStatus, error) {
	spec := &v1alpha1.KratosSpec{}
	status := &v1alpha1.KratosStatus{
		ScaleUpEvents:   make([]v1alpha1.ScaleChangeEvent, 0, 10),
		ScaleDownEvents: make([]v1alpha1.ScaleChangeEvent, 0, 10),
	}

	specAsString, found := data[specKey]

	if !found {
		return spec, status, errors.New("'kratosSpec' key not found.")
	}

	err := yaml.Unmarshal([]byte(specAsString), spec)

	if err != nil {
		return spec, status, err
	}

	statusAsString, found := data[statusKey]

	if !found {
		return spec, status, nil
	}

	err = yaml.Unmarshal([]byte(statusAsString), status)

	return spec, status, err
}

func (f *ScaleFacade) marshall(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus) (string, string, error) {

	specAsBytes, err := yaml.Marshal(spec)

	if err != nil {
		return empty, empty, err
	}

	statusAsBytes, err := yaml.Marshal(status)

	if err != nil {
		return empty, empty, err
	}

	return string(specAsBytes), string(statusAsBytes), nil
}

func (f *ScaleFacade) calculateMaxScaleReplicas(item *corev1.ConfigMap, currentReplicas int32, spec *v1alpha1.KratosSpec) int32 {
	log := f.log.WithValues("namespace", item.GetNamespace(), "name", item.GetName())
	maxReplicaProposal := int32(0)
	for _, metric := range spec.Metrics {
		metricFetcher, err := f.metricsFactory.GetMetricsFetcher(&metric)

		if err != nil {
			log.Error(err, "No fetcher defined for metric type", "type", metric.Type)
			f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "MetricFetcherTypeError", "No fetcher defined for metric type: %v", err.Error())
			continue
		}

		metricValues, err := metricFetcher.Fetch(&metric)

		if err != nil {
			log.Error(err, "error on fetching metric", "metric", metric)
			f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "MetricFetchError", "error on fetching metric: %v, error: %v", metric, err.Error())
			continue
		}
		replicaProposal, err := f.replicaCalculator.CalculateReplicas(currentReplicas, metric.Prometheus.Target, metricValues)

		log.V(1).Info("metric values and replica proposal", "replicas", replicaProposal, "metrics", metricValues)

		if err != nil {
			log.Error(err, "error on calculating replicas proposal for metric", "metric", metric)
			f.eventRecorder.Eventf(item, corev1.EventTypeWarning, "CalculateMetricReplicasError", "error on calculating replicas proposal for metric: %v, error: %v", metric, err.Error())
		}

		if maxReplicaProposal < replicaProposal {
			maxReplicaProposal = replicaProposal
		}
	}

	if maxReplicaProposal > spec.MaxReplicas {
		return spec.MaxReplicas
	}

	if maxReplicaProposal < spec.MinReplicas {
		return spec.MinReplicas
	}

	return maxReplicaProposal
}

func (f *ScaleFacade) updateObjects(originalItem *corev1.ConfigMap, spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus) error {
	log := f.log.WithValues("item", originalItem.GetName())
	_, statusAsString, err := f.marshall(spec, status)

	if err != nil {
		return err
	}

	log.V(1).Info("creating Patch...")

	patchedConfigMap := originalItem.DeepCopy()
	patchedConfigMap.Data["kratosStatus"] = statusAsString

	log.V(1).Info("updating on server")

	err = f.client.Patch(context.TODO(), patchedConfigMap, client.MergeFrom(originalItem))
	return err
}

func (f *ScaleFacade) expireRecommendationsAndScaleEvents(spec *v1alpha1.KratosSpec, status *v1alpha1.KratosStatus) {
	longestScaleUpWindow := int32(0)
	longestScaleDownWindow := int32(0)

	if spec.Behavior != nil {
		if spec.Behavior.ScaleUp != nil {
			longestScaleUpWindow = f.findLongestPolicyWindow(spec.Behavior.ScaleUp.Policies)
		}

		if spec.Behavior.ScaleDown != nil {
			longestScaleDownWindow = f.findLongestPolicyWindow(spec.Behavior.ScaleDown.Policies)
		}
	}

	maxStabilizationWindowSeconds := common.Max(spec.StabilizationWindowSeconds, common.Max(longestScaleUpWindow, longestScaleDownWindow))
	status.Recommendations = f.expireRecommendations(maxStabilizationWindowSeconds, status.Recommendations)

	status.ScaleUpEvents = f.expireEvents(longestScaleUpWindow, status.ScaleUpEvents)
	status.ScaleDownEvents = f.expireEvents(longestScaleDownWindow, status.ScaleDownEvents)
}

func (f *ScaleFacade) findLongestPolicyWindow(policies []v1alpha1.ScalingPolicy) int32 {
	if policies == nil || len(policies) == 0 {
		return 0
	}

	longestWindow := int32(0)

	for _, policy := range policies {
		if policy.PeriodSeconds > longestWindow {
			longestWindow = policy.PeriodSeconds
		}
	}

	return longestWindow
}

func (f *ScaleFacade) expireRecommendations(windowSeconds int32, recommendations []v1alpha1.Recommendation) []v1alpha1.Recommendation {
	result := recommendations[:0]
	cutOff := time.Now().Add(time.Duration(-windowSeconds))
	for _, recommendation := range recommendations {
		if recommendation.Timestamp.Time.After(cutOff) {
			result = append(result, recommendation)
		}
	}
	return result
}

func (f *ScaleFacade) expireEvents(windowSeconds int32, events []v1alpha1.ScaleChangeEvent) []v1alpha1.ScaleChangeEvent {
	result := events[:0]
	cutOff := time.Now().Add(time.Duration(-windowSeconds))
	for _, event := range events {
		if event.Timestamp.Time.After(cutOff) {
			result = append(result, event)
		}
	}
	return result
}

func (f *ScaleFacade) recordRecommendation(proposedReplicas int32, status *v1alpha1.KratosStatus) {
	recommendation := v1alpha1.Recommendation{Replicas: proposedReplicas, Timestamp: metav1.Now()}
	status.Recommendations = append(status.Recommendations, recommendation)
}

func (f *ScaleFacade) recordScaleEvent(currentReplicas int32, proposedReplicas int32, status *v1alpha1.KratosStatus) {
	if currentReplicas > proposedReplicas {
		status.ScaleDownEvents = append(status.ScaleDownEvents, v1alpha1.ScaleChangeEvent{Timestamp: metav1.Now(), ReplicaChange: currentReplicas - proposedReplicas})
	} else {
		status.ScaleUpEvents = append(status.ScaleUpEvents, v1alpha1.ScaleChangeEvent{Timestamp: metav1.Now(), ReplicaChange: proposedReplicas - currentReplicas})
	}
}
