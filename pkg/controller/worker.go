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
	"fmt"
	"time"

	"github.com/adobe/kratos/pkg/api/common"
	"github.com/adobe/kratos/pkg/scale"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DELETED     bool = true
	NOT_DELETED bool = false
)

type Worker struct {
	client      client.Client
	log         logr.Logger
	queue       workqueue.RateLimitingInterface
	scaleFacade *scale.ScaleFacade
}

func NewWorker(params *common.KratosParameters) (*Worker, error) {
	scaleFacade, err := scale.NewScaleFacade(params)

	if err != nil {
		return nil, err
	}

	worker := &Worker{
		client:      params.Client,
		log:         ctrl.Log.WithName("scale-worker"),
		queue:       workqueue.NewNamedRateLimitingQueue(NewFixedItemIntervalRateLimiter(10*time.Second), "kratosautoscaler"),
		scaleFacade: scaleFacade,
	}

	return worker, err
}

func (s *Worker) addItem(name types.NamespacedName) {
	s.queue.AddRateLimited(name)
}

func (s *Worker) removeItem(name types.NamespacedName) {
	s.queue.Forget(name)
	s.queue.Done(name)
}

func (s *Worker) run(threadiness int, stopCh chan struct{}) {
	// don't let panics crash the process
	defer utilruntime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	defer s.queue.ShutDown()

	for i := 0; i < threadiness; i++ {
		// runWorker will loop until "something bad" happens.  The .Until will
		// then rekick the worker after one second
		go wait.Until(s.runWorker, time.Second, stopCh)
	}

	// wait until we're told to stop
	<-stopCh
	s.log.Info("Shutting down Worker")

}

func (s *Worker) runWorker() {
	s.log.Info("Starting worker")
	for s.processNextWorkItem() {
	}
}

func (s *Worker) processNextWorkItem() bool {
	key, quit := s.queue.Get()
	if quit {
		return false
	}
	defer s.queue.Done(key)

	deleted, err := s.processItem(key.(types.NamespacedName))
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("%v failed with : %v", key, err))
	}

	if !deleted {
		s.queue.AddRateLimited(key)
	}

	return true
}

func (s *Worker) processItem(name types.NamespacedName) (deleted bool, err error) {
	s.log.Info("scaling ", "item", name)
	configMap := &corev1.ConfigMap{}

	errRetrieve := s.client.Get(context.TODO(), name, configMap)
	if errRetrieve != nil {
		if errors.IsNotFound(errRetrieve) {
			s.log.Info("ConfigMap not found. Ignoring since object must be deleted.", "item", name)

			return DELETED, nil
		}
		return NOT_DELETED, err
	}

	s.scaleFacade.Scale(configMap)
	return NOT_DELETED, nil
}
