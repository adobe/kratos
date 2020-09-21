/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/
package common

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KratosParameters struct {
	ClientConfig               *rest.Config
	Client                     client.Client
	RestMapper                 meta.RESTMapper
	EventRecorder              record.EventRecorder
	DefaultPrometheusUrl       string
	StabilizationWindowSeconds int32
}
