// Copyright 2020 Adobe
// All Rights Reserved.
//
// NOTICE: Adobe permits you to use, modify, and distribute this file in
// accordance with the terms of the Adobe license agreement accompanying
// it. If you have received this file from a source other than Adobe,
// then your use, modification, or distribution of it requires the prior
// written permission of Adobe.

package dockerutils

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"

	// mage:import utils
	"github.com/adobe/kratos/mage/util"
)

const (
	imageName = "kratos"
)

// Build docker image
func BuildImage() {
	cli := createClient()
	defer cli.Close()

	path, err := os.Getwd()

	utils.PanicOnError(err)

	ctx, _ := archive.TarWithOptions(path, &archive.TarOptions{})

	options := types.ImageBuildOptions{Tags: []string{createImageTag()}}

	response, err := cli.ImageBuild(context.TODO(), ctx, options)
	utils.PanicOnError(err)

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, os.Stdout.Fd(), true, nil)
	utils.PanicOnError(err)
}

// Push image to registry
func PushImage() {
	cli := createClient()
	defer cli.Close()

	response, err := cli.ImagePush(context.TODO(), createImageTag(), types.ImagePushOptions{})
	utils.PanicOnError(err)

	err = jsonmessage.DisplayJSONMessagesStream(response, os.Stdout, os.Stdout.Fd(), true, nil)
	utils.PanicOnError(err)
}

func createClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	utils.PanicOnError(err)
	return cli
}

func createImageTag() string {
	tag := utils.GetVersion()
	return fmt.Sprintf("%s:%s", imageName, tag)
}
