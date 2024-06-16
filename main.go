// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	updateProv "github.com/dr/packer-plugin-win-update/provisioner/update"
	updateVersion "github.com/dr/packer-plugin-win-update/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("my-provisioner", new(updateProv.Provisioner))
	pps.SetVersion(updateVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
