# packer2terraform

packer2terraform turns Packer's [machine-readable output](https://packer.io/docs/command-line/machine-readable.html) into [Terraform-readable tfvars](https://terraform.io/docs/configuration/variables.html). For example, you have Packer build an AMI that Terraform deploys to AWS.

[![travis build status for packer2terraform](https://travis-ci.org/saymedia/packer2terraform.svg)](https://travis-ci.org/saymedia/packer2terraform) [![Coverage Status](https://coveralls.io/repos/saymedia/packer2terraform/badge.svg?branch=master)](https://coveralls.io/r/saymedia/packer2terraform?branch=master)

## Usage

packer2terraform reads from STDIN and writes to STDOUT.

    packer2terraform -f [input filename] -template [template filename]

## Example

    packer -machine-readable build app.json | packer2terraform > app.tfvars
    
Or:

    packer2terraform -f packer_out.csv -template tmpl.tfvars > app.tfvars

Given this CSV input:

    1432168589,amazon-ebs,artifact-count,2
    1432168589,amazon-ebs,artifact,0,builder-id,mitchellh.amazonebs
    1432168589,amazon-ebs,artifact,0,id,us-west-1:ami-df76909b
    1432168589,amazon-ebs,artifact,0,string,AMIs were created:\n\nus-west-1: ami-df76909b
    1432168589,amazon-ebs,artifact,0,files-count,0
    1432168589,amazon-ebs,artifact,0,end
    1432168589,amazon-ebs,artifact,1,builder-id,mitchellh.amazonebs
    1432168589,amazon-ebs,artifact,1,id,us-west-2:ami-df79909c
    1432168589,amazon-ebs,artifact,1,string,AMIs were created:\n\nus-west-2: ami-df79909c
    1432168589,amazon-ebs,artifact,1,files-count,0
    1432168589,amazon-ebs,artifact,1,end

And this template:

    variable "images" {
        default = {
    {{range .Artifacts}}
            {{index .IDSplit 0}} = "{{index .IDSplit 1}}"{{end}}
        }
    }

packer2terraform will produce this output:

    variable "images" {
        default = {
    
            us-west-1 = "ami-df79909b"
            us-west-2 = "ami-df79909c"
        }
    }

## Install

    go get github.com/saymedia/packer2terraform

## Test

    go test ./...

Running `./test.sh` additionally tests using `go vet`, `golint`, `gocyclo`, `gofmt` and `go build`, which should be done before a commit.

## License

Copyright © 2015 Say Media Ltd. All Rights Reserved. See the LICENSE file for distribution terms.
