# Packer Plugin Windows Update

This repository is a template for a Packer multi-component plugin. It is intended as a starting point for creating Packer plugins, containing:
- A provisioner ([provisioner/update](provisioner/update))
- Docs ([docs](docs))
- A working example ([example](example))

These folders contain boilerplate code that you will need to edit to create your own Packer multi-component plugin.
A full guide to creating Packer plugins can be found at [Extending Packer](https://www.packer.io/docs/plugins/creation).

In this repository you will also find a pre-defined GitHub Action configuration for the release workflow
(`.goreleaser.yml` and `.github/workflows/release.yml`). The release workflow configuration makes sure the GitHub
release artifacts are created with the correct binaries and naming conventions.

## Build from source

1. Clone this GitHub repository locally.

2. Run this command from the root directory: 
```shell 
go build -ldflags="-X github.com/DrSkillIssue/packer-plugin-win-update/version.VersionPrerelease=dev" -o packer-plugin-win-update
```

3. After you successfully compile, the `packer-plugin-win-update` plugin binary file is in the root directory. 

4. To install the compiled plugin, run the following command 
```shell
packer plugins install --path packer-plugin-win-update github.com/DrSkillIssue/win-update
```

### Build on *nix systems
Unix like systems with the make, sed, and grep commands installed can use the `make dev` to execute the build from source steps. 

### Build on Windows Powershell
The preferred solution for building on Windows are steps 2-4 listed above.
If you would prefer to script the building process you can use the following as a guide

```powershell
$MODULE_NAME = (Get-Content go.mod | Where-Object { $_ -match "^module"  }) -replace 'module ',''
$FQN = $MODULE_NAME -replace 'packer-plugin-',''
go build -ldflags="-X $MODULE_NAME/version.VersionPrerelease=dev" -o packer-plugin-win-update.exe
packer plugins install --path packer-plugin-win-update.exe $FQN
```

# Requirements

-	[packer-plugin-sdk](https://github.com/hashicorp/packer-plugin-sdk) >= v0.5.2
-	[Go](https://golang.org/doc/install) >= 1.22.4

## Packer Compatibility
This update template is compatible with Packer >= v1.10.2
