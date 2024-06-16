// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package update

import (
	"context"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	// Default Restart Timeout
	DefaultRestartTimeout = "5m"

	// Default Upload Retry Attempts
	DefaultUploadRetryAttempts = 5

	// Default Upload Retry Delay
	DefaultUploadRetryDelay = 10 * time.Second

	// Default Upload Timeout
	DefaultUploadTimeout = 1 * time.Minute

	// Default Update Retry Attempts
	DefaultUpdateRetryAttempts = 5

	// Specifies the restart command (if required).
	RestartCommand = "shutdown.exe -f -r -t 0 -c \"Packer Windows Update Restart\""

	// Specifies the default action to install all updates if no choices were provided.
	InstallAllUpdates = true
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Specifies the username to use for performing Windows Update.
	Username string `mapstructure:"username"`

	// Specifies the password to use for performing Windows Update.
	Password string `mapstructure:"password"`

	// Specifies Windows Update Category IDs (NOTE: these are UUIDs, not names).
	CategoryIDs []string `mapstructure:"category_ids"`

	// Specifies one to many cab files to install.
	// NOTE: These are locations where the cab files are stored on the machine.
	CabFiles []string `mapstructure:"cab_files"`

	// Specifies a flag to install all updates.
	// NOTE: To include hidden updates, set the include_hidden flag to true.
	InstallAllUpdates bool `mapstructure:"install_all"`

	// Specifies a flag to include hidden updates.
	IncludeHiddenUpdates bool `mapstructure:"include_hidden"`

	// Specifies a flag to install optional updates.
	// NOTE: Optional updates are included via 'install_all' flag.
	InstallOptionalUpdates bool `mapstructure:"install_optional"`

	// Specifies a flag to install recommended updates.
	// NOTE: Recommended updates are included via 'install_all' flag.
	InstallRecommendedUpdates bool `mapstructure:"install_recommended"`

	// Specifies a flag to install important updates.
	// NOTE: Important updates are included via 'install_all' flag.
	InstallImportantUpdates bool `mapstructure:"install_important"`

	// Specifies
	ctx interpolate.Context
}

type Provisioner struct {
	config       Config
	communicator packer.Communicator
	ui           packer.Ui
	cancel       chan struct{}
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	if err != nil {
		return err
	}

	return nil
}

//func (p *Provisioner) Provision(_ context.Context, ui packer.Ui, _ packer.Communicator, generatedData map[string]interface{}) error {
//	ui.Say(fmt.Sprintf("provisioner mock: %s", p.config.MockOption))
//	return nil
//}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, _ map[string]interface{}) error {
	p.communicator = comm
	p.ui = ui
	p.cancel = make(chan struct{})

	/* err = retry.Config{StartTimeout: uploadTimeout}.Run(ctx, func(context.Context) error {
		if err := comm.Upload(
			elevatedPath,
			bytes.NewReader(buffer.Bytes()),
			nil); err != nil {
			return fmt.Errorf("Error uploading the Windows update elevated script: %s", err)
		}
		return nil
	})
	if err != nil {
		return err
	} */
	return nil
}
