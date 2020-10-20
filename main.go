/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package main

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/adobe/kratos/api/common"
	"github.com/adobe/kratos/controllers"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	scalingv1alpha1 "github.com/adobe/kratos/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// -- int32 Value
type int32Value int32

func newInt32Value(val int32, p *int32) *int32Value {
	*p = val
	return (*int32Value)(p)
}

func (i *int32Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		err = numError(err)
	}
	*i = int32Value(v)
	return err
}

func (i *int32Value) Get() interface{} { return int32(*i) }

func (i *int32Value) String() string { return strconv.FormatInt(int64(*i), 10) }

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errors.New("parse error")
	}
	if ne.Err == strconv.ErrRange {
		return errors.New("value out of range")
	}
	return err
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(scalingv1alpha1.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var namespacesList string
	var defaultPrometheusUrl string
	var defaultStabilizationWindowSeconds int32

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&namespacesList, "namespaces", "", "Comma separated list of namespaces")
	flag.StringVar(&defaultPrometheusUrl, "default-prometheus-url", "https://prometheus-monitoring-va7.int.pipeline.adobedc.net", "Default Prometheus url")
	flag.Var(newInt32Value(300, &defaultStabilizationWindowSeconds), "stabilization-window-seconds", "Stabilization window in seconds")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.Level(zapcore.DebugLevel)))

	namespaces := strings.Split(namespacesList, ",")
	setupLog.Info("Listening for namespaces", "namespaces", namespaces)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "a9cf7578.core.adobe.com",
		NewCache:           cache.MultiNamespacedCacheBuilder(namespaces),
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	params := &common.KratosParameters{
		ClientConfig:               mgr.GetConfig(),
		Client:                     mgr.GetClient(),
		RestMapper:                 mgr.GetRESTMapper(),
		EventRecorder:              mgr.GetEventRecorderFor("kratos"),
		DefaultPrometheusUrl:       defaultPrometheusUrl,
		StabilizationWindowSeconds: defaultStabilizationWindowSeconds,
	}

	reconciler, err := controllers.NewKratosReconciler(params)

	if err != nil {
		setupLog.Error(err, "unable to create reconciler")
		os.Exit(1)
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup controller with manager", "controller", "Kratos")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
