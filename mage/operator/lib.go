// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.

package operator

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	// mage:import utils
	"github.com/adobe/kratos/mage/util"
)

const (
	controllerGen = "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.3"
	buildDir      = "build/bin"
)

// Cleans operator build dir
func Clean() {
	fmt.Printf("- Clean operator build dir: %s\n", buildDir)
	sh.Rm(buildDir)
}

// Builds operator code
func Build() {
	mg.SerialDeps(Clean, GenerateDeepCopy, Vet, Fmt)
	fmt.Println("- Build")
	err := sh.RunV("go", "build", "-o", "bin/manager", "main.go")
	utils.PanicOnError(err)
}

// Runs operator agains configured kube config
func Run() {
	namespaces := os.Getenv("KRATOS_NAMESPACES")
	mg.SerialDeps(Clean, GenerateDeepCopy, Vet, Fmt, GenerateCrds)
	fmt.Println("- Running operator, namespaces: ", namespaces)
	err := sh.RunV("go", "run", "main.go", "--namespaces="+namespaces)
	utils.PanicOnError(err)
}

// Formats Go source files
func Fmt() {
	fmt.Println("- Fmt")
	path, err := os.Getwd()
	utils.PanicOnError(err)
	err = sh.RunV("go", "fmt", path+"/...")
	utils.PanicOnError(err)
}

// Runs static analysis of the code
func Vet() {
	fmt.Println("- Vet")
	path, err := os.Getwd()
	utils.PanicOnError(err)
	err = sh.RunV("go", "vet", path+"/...")
	utils.PanicOnError(err)
}

// Generates DeepCopy code using as file header "hack/boilerplate.go.txt"
func GenerateDeepCopy() {
	mg.Deps(installControllerGen)
	fmt.Println("- Generating DeepCopy code")
	err := sh.RunV("controller-gen", "object:headerFile=\"hack/boilerplate.go.txt\"", "paths=\"./...\"")
	utils.PanicOnError(err)
}

// Generates CRDs
func GenerateCrds() {
	mg.Deps(installControllerGen)
	fmt.Println("- Generating CRDs")
	err := sh.RunV("controller-gen", "crd:trivialVersions=true", "paths=\"./...\"", "output:crd:artifacts:config=config/crd")
	utils.PanicOnError(err)
}

// Run operator tests
func Test() {
	mg.Deps(Build)
	fmt.Println("- Running tests")
	err := sh.RunV("go", "test", "./...", "-coverprofile", "cover.out", "-test.v", "-ginkgo.v")
	utils.PanicOnError(err)

}

// Shows test coverage report
func CoverageReport() {
	err := sh.RunV("go", "tool", "cover", "-func", "cover.out")
	utils.PanicOnError(err)
}

func installControllerGen() {
	installBinary("controller-gen", controllerGen)
}

func installBinary(binaryName string, binarySrc string) {
	fmt.Printf("Checking if %s is installed\n", binaryName)
	var _, err = exec.LookPath(binaryName)

	if err == nil {
		fmt.Printf("%s - OK\n", binaryName)
		return
	}

	fmt.Printf("%s not installed\n", binaryName)
	err = sh.RunV("go", "get", binarySrc)
	utils.PanicOnError(err)
}
