// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.

package helm

import (
	"fmt"
	"path/filepath"

	utils "github.com/adobe/kratos/mage/util"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"helm.sh/helm/v3/pkg/action"
)

const buildDir = "build/helm"

// Cleans helm build directory
func Clean() {
	fmt.Printf("- Cleaning build dir: %s\n", buildDir)
	sh.Rm(buildDir)
}

// Packages kratos-operator helm chart
func PackageChart() {
	mg.Deps(Clean)
	fmt.Println("- Packaging chart")
	fmt.Printf("- Destination %s\n", buildDir)
	path, err := filepath.Abs("helm/kratos-operator")
	utils.PanicOnError(err)

	client := action.NewPackage()
	client.Destination = buildDir
	result, err := client.Run(path, nil)
	utils.PanicOnError(err)
	fmt.Printf("Successfully packaged chart and saved it to: %s\n", result)
}
