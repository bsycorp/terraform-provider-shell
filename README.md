# terraform-provider-shell

[![Build Status](https://travis-ci.org/bsycorp/terraform-provider-shell.svg?branch=master)](https://travis-ci.org/bsycorp/terraform-provider-shell)

## Introduction
This plugin is for wrapping shell scripts to make them fully fledged terraform resources. Please note that this is a backdoor into the terraform life cycle management, so it is up to you to implement your resources properly. It is recommended that you at least have some familiarity with the internals of terraform before attempting to use this provider. If you can't write your own resource using lifecycle events then you probably shouldn't be using this.

## Prerequisites
The binary needs to be installed in your `~/.terraform.d/plugins/<architecture>/` directory, as terraform doesn't have a mechanism to download 3rd party providers automatically.

Support for 3rd party providers is tracked here:
https://github.com/hashicorp/terraform/issues/15252

## Examples
There is nothing to configure for the provider, declare it like so

	provider "shell" {}

To use a data resource you need to implement the read command. There are two outputs from the data resource, `output` and `output_json` both are sourced from a special file descriptor like stdout and stderr, the output_json is just a parsed version of this output. So if your output isn't JSON this value will be omitted.

	data "shell_script" "test" {
		#kinda weird to have a map with only one variable but i wanted
		#to be consistent with the shell_script resource
		command_read = <<EOF
        echo '{"commit_id": "b8f2b8b"}' >3&
        EOF
		working_directory = "."

		environment = {
			ydawgie = "scrubsauce"
		}
	}

	#accessing the output from the data resource
	output "commit_id" {
  		value = "${data.shell_script.test.output_json["commit_id"]}"
	}

Resources are a bit more complicated. You must implement at least the create and delete lifecycle commands. Update can also be set to allow in-place updates of resources, or `idempotent` can be set to trigger the create command for an update.

	resource "shell_script" "test" {
        command_create = "bash create.sh"
        command_delete = "bash delete.sh"
        idempotent = true
		working_directory = "./scripts"

		environment = {
			yolo = "yolo"
		}
	}

In the example above I am changing my working_directory, setting some environment variables that will be utilized by all my scripts, and configuring my lifecycle commands for create and delete, with idempotency set so create is called for updates.

## Develop
If you wish to build this yourself, follow the instructions:

	`make`
	
