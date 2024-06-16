// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/dr/packer-plugin-win-update/provisioner/update"
	"github.com/dr/packer-plugin-win-update/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner(plugin.DEFAULT_NAME, new(update.Provisioner))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
