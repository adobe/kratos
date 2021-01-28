// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.
//
//+build mage

package main

import (
	"github.com/magefile/mage/mg"

	// mage:import docker
	"github.com/adobe/kratos/mage/docker"

	// mage:import helm
	"github.com/adobe/kratos/mage/helm"

	// mage:import operator
	"github.com/adobe/kratos/mage/operator"

	// mage:import integration_tests
	"github.com/adobe/kratos/mage/integration_tests"
)

// Cleans project
func Clean() {
	//mg.Deps(operator.Clean)
	mg.Deps(operator.Clean, helm.Clean)
}

// Build code, docker image and helm chart
func Build() {
	mg.SerialDeps(operator.Build)
	mg.SerialDeps(helm.PackageChart)
	mg.SerialDeps(dockerutils.BuildImage)
	mg.SerialDeps(dockerutils.PushImage)
}

// Run Integration Tests: `mage RunIntegrationTests`
func RunIntegrationTests() {
	mg.SerialDeps(integration_tests.Run)
}
