#!/bin/bash
VERSION="0.1.0"
echo "Downloading plugin to ~/.terraform.d/plugins/.."
mkdir -p ~/.terraform.d/plugins/darwin_amd64
curl -sSL https://github.com/bsycorp/terraform-provider-shell/releases/download/$VERSION/terraform-provider-shell_v$VERSION-darwin-amd64 -o ~/.terraform.d/plugins/darwin_amd64/terraform-provider-shell_v$VERSION
chmod +x ~/.terraform.d/plugins/darwin_amd64/terraform-provider-shell_v$VERSION

mkdir -p ~/.terraform.d/plugins/linux_amd64
curl -sSL https://github.com/bsycorp/terraform-provider-shell/releases/download/$VERSION/terraform-provider-shell_v$VERSION-linux-amd64 -o ~/.terraform.d/plugins/linux_amd64/terraform-provider-shell_v$VERSION
chmod +x ~/.terraform.d/plugins/linux_amd64/terraform-provider-shell_v$VERSION
echo "Done"
