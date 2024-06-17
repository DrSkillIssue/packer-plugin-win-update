// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package update

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf16"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

const (
	// Specifies the amount of time to wait for a restart after updates are applied.
	DefaultRestartTimeout = 1 * time.Hour

	// Specifies the amount of times to retry uploading the Windows Update script.
	DefaultUploadRetryAttempts = 5

	// Specifies the delay between retries to upload the Windows Update script.
	DefaultUploadRetryDelay = 30 * time.Second

	// Specifies the amount of time to wait for new Windows Update script to be uploaded.
	DefaultUploadTimeout = 5 * time.Minute

	// Specifies the amount of times to invoke the Windows Update script (if necessary).
	DefaultUpdateRetryAttempts = 5

	// Specifies the default time to wait for Windows to finish updating.
	DefaultUpdateTimeout = 4 * time.Hour

	// Specifies the default time to wait before retrying the Windows Update script.
	DefaultUpdateRetryDelay = 10 * time.Second

	// Default Update Attempts
	// An attempt is defined as running the Windows Update script.
	DefaultUpdateAttempts = 3

	// Specifies the restart command (if required).
	RestartCommand = "shutdown.exe -f -r -t 0 -c \"Packer Windows Update Restart\""
)

// Populate 'script' with the contents of the Windows Update script.
//
//go:embed Invoke-WinUpdate.ps1
var script []byte

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context

	// Specifies the username to use for performing Windows Update.
	Username string `mapstructure:"username" required:"false"`

	// Specifies the password to use for performing Windows Update.
	Password string `mapstructure:"password" required:"false"`

	// Specifies Windows Update Category IDs (NOTE: these are UUIDs, not names).
	CategoryIDs []string `mapstructure:"category_ids" required:"false"`

	// Specifies one to many cab files to install.
	// NOTE: These are locations where the cab files are stored on the machine.
	CabFiles []string `mapstructure:"cab_files" required:"false"`

	// Specifies a flag to install all updates.
	// NOTE: To include hidden updates, set the include_hidden flag to true.
	InstallAllUpdates bool `mapstructure:"install_all" required:"false"`

	// Specifies a flag to include hidden updates.
	IncludeHiddenUpdates bool `mapstructure:"include_hidden" required:"false"`

	// Specifies a flag to install optional updates.
	// NOTE: Optional updates are included via 'install_all' flag.
	InstallOptionalUpdates bool `mapstructure:"install_optional" required:"false"`

	// Specifies a flag to install recommended updates.
	// NOTE: Recommended updates are included via 'install_all' flag.
	InstallRecommendedUpdates bool `mapstructure:"install_recommended" required:"false"`

	// Specifies a flag to install important updates.
	// NOTE: Important updates are included via 'install_all' flag.
	InstallImportantUpdates bool `mapstructure:"install_important" required:"false"`

	// Specifies the update timeout.
	// NOTE: If not specified, the default is 4 hours.
	UpdateTimeout time.Duration `mapstructure:"update_timeout" required:"false"`

	// Specifies the amount of times to retry updates.
	// NOTE: If not specified, the default is 3 attempts.
	UpdateRetryAttempts uint `mapstructure:"update_retry_attempts" required:"false"`

	// Specifies the number of retries to attempt uploading the script.
	// NOTE: If not specified, the default is 5 attempts.
	UploadRetryAttempts uint `mapstructure:"upload_retry_attempts" required:"false"`

	// Specifies the delay between retries to upload the script.
	// NOTE: If not specified, the default is 30 seconds.
	UploadRetryDelay time.Duration `mapstructure:"upload_retry_delay" required:"false"`

	// Specifies the timeout to upload the script.
	// NOTE: If not specified, the default is 5 minutes.
	UploadTimeout time.Duration `mapstructure:"upload_timeout" required:"false"`

	// Specifies the amount of time to wait for a restart.
	// NOTE: If not specified, the default is 1 hour.
	RestartTimeout time.Duration `mapstructure:"restart_timeout" required:"false"`

	// Specifies whether to disable restarting after updates.
	// NOTE: If not specified, the default is false.
	DisableRestartAfterUpdates bool `mapstructure:"disable_restart" required:"false"`
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

	// Apply default values if not specified.
	if p.config.UploadRetryAttempts <= 0 {
		p.config.UploadRetryAttempts = DefaultUploadRetryAttempts
	}

	// Apply default upload retry dedlay if not specified.
	if p.config.UploadRetryDelay <= 0 {
		p.config.UploadRetryDelay = DefaultUploadRetryDelay
	}

	// Apply default upload timeout if not specified.
	if p.config.UploadTimeout <= 0 {
		p.config.UploadTimeout = DefaultUploadTimeout
	}

	// Apply default RestartTimeout if not specified.
	if p.config.RestartTimeout <= 0 {
		p.config.RestartTimeout = DefaultRestartTimeout
	}

	// Apply default UpdateRetryAttempts if not specified.
	if p.config.UpdateRetryAttempts <= 0 {
		p.config.UpdateRetryAttempts = DefaultUpdateRetryAttempts
	}

	// Apply default UpdateTimeout if not specified.
	if p.config.UpdateTimeout <= 0 {
		p.config.UpdateTimeout = DefaultUpdateTimeout
	}

	// Specifies whether user specified a specific update/updates.
	// If this results as 'false', we will install all updates.
	var specifiedUpdates bool = (p.config.InstallImportantUpdates ||
		p.config.InstallOptionalUpdates ||
		p.config.InstallRecommendedUpdates ||
		p.config.CabFiles != nil ||
		p.config.CategoryIDs != nil)

	// No update specified. Install all updates.
	if !specifiedUpdates {
		p.config.InstallAllUpdates = true
	}

	// Accumulate any errors in the configuration.
	var errs *packer.MultiError

	// Validate CabFiles are valid file paths.
	if p.config.CabFiles != nil {
		for _, cabFile := range p.config.CabFiles {
			if _, err := filepath.Abs(cabFile); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("invalid file path for provided cab file: %s", cabFile))
			}
		}
	}

	// Validate CategoryIDs are valid UUIDs.
	if p.config.CategoryIDs != nil {
		for _, categoryID := range p.config.CategoryIDs {
			if _, err := uuid.Parse(categoryID); err != nil {
				errs = packer.MultiErrorAppend(errs,
					fmt.Errorf("invalid Windows Category UUID format: %s", categoryID))
			}
		}
	}

	// Return errors if any.
	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

//func (p *Provisioner) Provision(_ context.Context, ui packer.Ui, _ packer.Communicator, generatedData map[string]interface{}) error {
//	ui.Say(fmt.Sprintf("provisioner mock: %s", p.config.MockOption))
//	return nil
//}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, _ map[string]interface{}) error {

	// Populate provisioner fields
	p.communicator = comm
	p.ui = ui
	p.cancel = make(chan struct{})

	// Initialize an empty filePath string.
	// If the script is successfully uploaded, this will be populated.
	var filePath string

	ui.Say("Starting Windows Update Provisioner.")

	// Upload the Windows Update script.
	filePath, err := p.uploadScript(ctx, ui, comm)

	// Error - Creating the Windows Update script.
	if err != nil {
		return err
	}

	// Error - File path is empty.
	if len(filePath) == 0 {
		return fmt.Errorf("expected file path to be populated after uploading the Windows Update script, but it was empty")
	}

	// Successfully uploaded the Windows Update script.
	ui.Say("Successfully created Windows Update script: " + filePath)

	// Iniitalize variable to store the results of Windows Update.
	var statusCode int

	// Run the Windows Update script.
	statusCode, err = p.runWindowsUpdate(ctx, ui, comm, filePath)
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

	if err != nil {
		return err
	}

	if statusCode != 0 {
		return fmt.Errorf("Windows Update script exited with non-zero exit status: %d", statusCode)
	}

	return nil
}

func (p *Provisioner) uploadScript(ctx context.Context, ui packer.Ui, comm packer.Communicator) (string, error) {

	ui.Say("Creating/Uploading Windows Update script to specified target.")

	// Error - Windows Update Script is empty.
	if len(script) == 0 {
		return "", fmt.Errorf("contents within 'Invoke-WinUpdate.ps1' are empty")
	}

	// If successful, this will store the file path of the newly created Windows Update script.
	var filePath string

	// Upload the Windows Update script and retry if necessary.
	err := retry.Config{
		StartTimeout: p.config.UploadTimeout,
		RetryDelay:   func() time.Duration { return p.config.UploadRetryDelay },
		Tries:        int(p.config.UploadRetryAttempts),
		ShouldRetry:  nil}.Run(ctx, func(context.Context) error {

		// Create a temporary file to store the Windows Update script.
		// The purpose of doing it this way is to avoid making assumption of what the 'temp' directory is.
		tmpFile, err := os.CreateTemp("", "Invoke-WinUpdate*.ps1")

		// Error - Creating temp file
		if err != nil {
			return fmt.Errorf("error preparing elevated shell script: %s", err)
		}

		// Store newly created file path.
		filePath = tmpFile.Name()

		// Next - Store contents of the Windows Update script into fileName.
		// First - Create a new buffer containing the Windows Update script's bytes.
		writer := bytes.NewBuffer(script)

		// Next - Write the contents of the Windows Update script to the temp file.
		if _, err := writer.WriteTo(tmpFile); err != nil {
			// Error - Writing content to temp file.
			return fmt.Errorf("error writing content to temp file: %s", err)
		}

		// Close the temp file.
		if err := tmpFile.Close(); err != nil {
			// Error - Closing temp file.
			return fmt.Errorf("error closing temp file: %s", err)
		}

		return nil
	})

	return filePath, err
}

func (p *Provisioner) runWindowsUpdate(ctx context.Context, ui packer.Ui, comm packer.Communicator, script string) (int, error) {

	ui.Say("Running Windows update...")

	// Error - File path is empty.
	if len(script) == 0 {
		return 1, fmt.Errorf("unable to run Windows Update script: file path is empty")
	}

	// After important updates, it's possible to have a pending restart.
	// If one is pending, this will be set to true.
	//var restartPending bool

	// Build the Windows Update command.
	command := p.getWindowsUpdateCommand(script)

	// Initialize a buffer to store the output of the Windows Update script.
	var scriptOutput bytes.Buffer

	// Initialize a buffer to store the errors of the Windows Update script.
	var scriptErrors bytes.Buffer

	// Initialize a variable to store the exit status of the Windows Update script.
	var exitStatus int

	err := retry.Config{
		StartTimeout: p.config.UpdateTimeout,
		Tries:        int(p.config.UpdateRetryAttempts),
		ShouldRetry:  nil,
	}.Run(ctx, func(context.Context) error {

		// Prepare the command to run the Windows Update script.
		cmd := &packer.RemoteCmd{
			Command: command,
			Stdout:  &scriptOutput,
			Stderr:  &scriptErrors,
		}

		// Run the Windows Update script.
		err := cmd.RunWithUi(ctx, comm, ui)

		if err != nil {
			return err
		}

		exitStatus = cmd.ExitStatus()

		switch exitStatus {
		case 0:
			return nil
		case 101:
			//restartPending = true
			return nil
		case 2147942501: //windows 2012
			//restartPending = true
			return nil
		default:
			return fmt.Errorf("Windows update script exited with non-zero exit status: %d", exitStatus)
		}
	})

	return exitStatus, err
}

func (p *Provisioner) getWindowsUpdateCommand(script string) string {

	return fmt.Sprintf(
		"PowerShell -ExecutionPolicy Bypass -NoProfile -OutputFormat Text -EncodedCommand %s",
		base64.StdEncoding.EncodeToString(encodeUtf16Le(script)))
}

func encodeUtf16Le(s string) []byte {

	d := utf16.Encode([]rune(s))

	b := make([]byte, len(d)*2)

	for i, r := range d {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}

	return b
}
