# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

locals {
  foo = data.update-my-datasource.mock-data.foo
  bar = data.update-my-datasource.mock-data.bar
}