#!/bin/bash
VERSION="0.1.0"
mkdir -p ~/.terraform.d/plugins/darwin_amd64
(cd ~/.terraform.d/plugins/darwin_amd64; curl https://github.com/bsycorp/terraform-provider-shell/releases/download/$VERSION/terraform-provider-shell_v$VERSION-darwin-amd64 -o terraform-provider-shell_v$VERSION)

mkdir -p ~/.terraform.d/plugins/linux_amd64
(cd ~/.terraform.d/plugins/linux_amd64; curl https://github.com/bsycorp/terraform-provider-shell/releases/download/$VERSION/terraform-provider-shell_v$VERSION-linux-amd64 -o terraform-provider-shell_v$VERSION)