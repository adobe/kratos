// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.

package integrationTests

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	harness "github.com/kudobuilder/kuttl/pkg/apis/testharness/v1beta1"
	"github.com/kudobuilder/kuttl/pkg/test"
	testutils "github.com/kudobuilder/kuttl/pkg/test/utils"
	// mage:import utils
)

// Runs Integration Tests using KUTTL
func RunIntegrationTests() error {

	options := harness.TestSuite{}

	configPath := "tests/integration/integrationTests.yaml"
	testToRun := "test"

	objects, err := testutils.LoadYAMLFromFile(configPath)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		kind := obj.GetObjectKind().GroupVersionKind().Kind

		if kind == "TestSuite" {
			options = *obj.(*harness.TestSuite)
		} else {
			log.Println(fmt.Errorf("unknown object type: %s", kind))
		}
	}

	testutils.RunTests("kuttl", testToRun, options.Parallel, func(t *testing.T) {
		harness := test.Harness{
			TestSuite: options,
			T:         t,
		}

		s, _ := json.MarshalIndent(options, "", "  ")
		fmt.Printf("Running integration tests with following options:\n%s\n", string(s))

		harness.Run()
	})

	return nil
}
