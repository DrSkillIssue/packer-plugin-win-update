<!--
  Include a short overview about the plugin.

  This document is a great location for creating a table of contents for each
  of the components the plugin may provide. This document should load automatically
  when navigating to the docs directory for a plugin.

-->

### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    name = {
      # source represents the GitHub URI to the plugin repository without the `packer-plugin-` prefix.
      source  = "github.com/organization/name"
      version = ">=0.0.1"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/organization/plugin-name
```

### Components

The update plugin is intended as a starting point for creating Packer plugins

#### Builders

- [builder](/packer/integrations/hashicorp/update/latest/components/builder/builder-name) - The update builder is used to create endless Packer
  plugins using a consistent plugin structure.

#### Provisioners

- [provisioner](/packer/integrations/hashicorp/update/latest/components/provisioner/provisioner-name) - The update provisioner is used to provisioner
  Packer builds.

#### Post-processors

- [post-processor](/packer/integrations/hashicorp/update/latest/components/post-processor/postprocessor-name) - The update post-processor is used to
  export update builds.

#### Data Sources

- [data source](/packer/integrations/hashicorp/update/latest/components/datasource/datasource-name) - The update data source is used to
  export update data.

