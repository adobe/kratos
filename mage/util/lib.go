// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.

package utils

import "os"

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func GetVersion() string {
	val := os.Getenv("PROJECT_VERSION")

	if val == "" {
		return "0.0.1"
	}

	return val
}
