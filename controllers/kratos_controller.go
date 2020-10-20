/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/
package controllers

import (
	"context"
	"time"

	"github.com/adobe/kratos/api/common"
	scalingv1alpha1 "github.com/adobe/kratos/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

// KratosReconciler reconciles a Kratos object
type KratosReconciler struct {
	client.Client
	log           logr.Logger
	queue         workqueue.RateLimitingInterface
	scalingWorker *Worker
	stopChan      chan struct{}
}

func NewKratosReconciler(params *common.KratosParameters) (*KratosReconciler, error) {
	scalingWorker, err := NewWorker(params)

	if err != nil {
		return nil, err
	}

	kratosReconciler := &KratosReconciler{
		Client:        params.Client,
		log:           ctrl.Log.WithName("reconciler"),
		queue:         workqueue.NewNamedRateLimitingQueue(NewFixedItemIntervalRateLimiter(10*time.Second), "kratosautoscaler"),
		scalingWorker: scalingWorker,
		stopChan:      make(chan struct{}),
	}
	go kratosReconciler.scalingWorker.run(1, kratosReconciler.stopChan)
	return kratosReconciler, nil
}

// +kubebuilder:rbac:groups=scaling.core.adobe.com,resources=kratos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.core.adobe.com,resources=kratos/status,verbs=get;update;patch
func (r *KratosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	name := req.NamespacedName
	log := r.log.WithValues("name", name)

	configMap := &corev1.ConfigMap{}

	err := r.Get(context.TODO(), req.NamespacedName, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigMap not found. Ignoring since object must be deleted.", "item", req.NamespacedName)
			r.scalingWorker.removeItem(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	if _, found := configMap.Data["kratosSpec"]; found {
		log.Info("ConfigMap has key 'kratosSpec', adding to queue.")
		r.scalingWorker.addItem(req.NamespacedName)
	} else {
		log.Info("ConfigMap missing 'kratosSpec', skipping.")
		r.scalingWorker.removeItem(req.NamespacedName)
	}

	return ctrl.Result{}, nil
}

func (r *KratosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalingv1alpha1.Kratos{}).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
