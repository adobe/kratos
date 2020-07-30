/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scalingv1alpha1 "github.com/adobe/kratos/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// KratosReconciler reconciles a Kratos object
type KratosReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=scaling.core.adobe.com,resources=kratos,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.core.adobe.com,resources=kratos/status,verbs=get;update;patch

func (r *KratosReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("kratos", req.NamespacedName)

	// your logic here
	configMap := &corev1.ConfigMap{}

	err := r.Get(context.Background(), req.NamespacedName, configMap)
	if err != nil {
		r.Log.Info("Deleted already")
	}
	r.Log.Info(configMap.ResourceVersion)

	return ctrl.Result{}, nil
}

func (r *KratosReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scalingv1alpha1.Kratos{}).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
