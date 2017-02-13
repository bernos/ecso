# ecso
ecso is a command line tool that provides high-level commands for building, 
deploying, running and monitoring projects on Amazon ECS. It's features include
- Create and manage multiple project "environments", such as production, 
  staging etc...
- Out of the box setup of cloudwatch logs, service discovery using route53 DNS, 
  CloudWatch alarms
- Opt-in support for running datadog on all container instances in an 
  environment
- Develop and deploy services using regular Docker Compose files
- Simple configuration via `.ecso/project.json`
- No magic - ecso creates and outputs garden variety CloudFormation templates 
  for everything under the hood. All ecso CloudFormation templates can be 
  freely modified by hand, and can be deployed using tools other than ecso, 
  such as the AWS cli or web console.
- Don't want a monolithic repository with all your instructure and service in 
  one place? ecso projects can easily span multiple repos: keep all your 
  environment infrastructure in one repository, and each of your services in 
  their own.

## Installing
If you have a working go environment, just run 
`go install github.com/bernos/ecso`. Otherwise, download the appropriate binary 
from the releases page, and add it to your `$PATH`

## Quick start

```bash
# Create a folder for your project
mkdir ~/my-project && cd ~/my-project

# Initialise a new ecso project. A project configuration file will be created 
# at .ecso/project.json. This file can safely be edited by hand after it is 
# created.
ecso init

# Set up a new ecso environment to deploy to. You'll be prompted for all the
# required details, such as VPC, subnets and so forth.
ecso environment add my-environment

# Now, create the resources for your new environment in AWS. For details of 
# what is created, see the cloudformation templates that ecso generates at 
# .ecso/infrastructure/templates. These cloudformation templates can also be 
# safely edited by hand, to customise your ecso infrastructure.
ecso environment up my-environment

# Create a new ecso service. This will update .ecso/project.json with 
# configuration settings for the service, as well as create a basic 
# docker-compose file at ./services/my-service, and a cloudforamtion template 
# at .ecso/services/my-service. Both the cloudforamtion templates and 
# docker-compose file can safely be edited by hand, in order to customise the 
# service or supporting resources
ecso service add my-service

# Now deploy the service to your environment
ecso service up my-service --environment my-environment

# Once the service is deployed, you can see the currently running services with
ecso service ls --environment my-environment

# List the containers running in the service with
ecso service ps my-service --environment my-environment

# You can view the logs of your running service with
ecso service logs my-service --environment my-environment

# Finally, to stop all running services, and destory the environment run
ecso environment rm my-environment --force
```

## Configuration defaults
Remembering VPC and subnet IDs and other details can be annoying. You can store
per-account settings for ecso environments in ~/.ecso.json and ecso will use
these settings as defaults, based on AWS account you are signed into when 
running ecso. Just copy, paste and edit the following json to ~/.ecso.json

```json
{
    "AccountDefaults": {
        "aws-numeric-account-id-here": {
            "VPCID": "vpc-abc123",
            "ALBSubnets": "subnet-abc123,subnet-abc456,subnet-abc789",
            "InstanceSubnets": "subnet-def123,subnet-def456,subnet-def789",
            "DNSZone": "connect.dev.outfra.xyz",
            "DataDogAPIKey": "sjkado9dfajkcoisn"
        },
        "another-aws-numeric-account-id-here": {
            "VPCID": "vpc-abc123",
            "ALBSubnets": "subnet-abc123,subnet-abc456,subnet-abc789",
            "InstanceSubnets": "subnet-def123,subnet-def456,subnet-def789",
            "DNSZone": "connect.dev.outfra.xyz",
            "DataDogAPIKey": "sjkado9dfajkcoisn"
        }
    }
}
```

# Building ecso

## Developing
Dependencies are vendored in the usual way, but managed with godep. Make sure 
you have the lates version of godep installed by running 
`go get -u github.com/tools/godep` When adding new dependencies, be sure to run 
`godep save ./...` in the project root.

## Building
Run `make build`. The resulting binary will be output at `bin/local`

## Testing
Run `make test` to run unit tests

## Releasing
- First, make sure to increment the `VERSION` number in `./Makefile`
- Ensure that the `GITHUB_TOKEN` env var is set to a github personal access 
  token that has write access to the ecso repository on github
- Run `make release`. This will create a git tag as well as a github release, 
  and updload binaries for all supported platforms to the github release
