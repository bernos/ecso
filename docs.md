
## init

Initialise a new ecso project

Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.

````
main init [PROJECT]
````
# ecso environment

Manage ecso environments

````
ecso environment command [command options] [arguments...]
````

#### Commands
- add		Add a new environment to the project
- up		Deploys the infrastructure for an ecso environment
- rm		Removes an ecso environment
- describe	Describes an ecso environment
- down		Terminates an ecso environment


#### Options
   --help, -h	show help
   
## add

Add a new environment to the project

````
ecso environment add [command options] [ENVIRONMENT]
````

#### Options
- --vpc value			The vpc to create the environment in
- --alb-subnets value		The subnets to place the application load balancer in
- --instance-subnets value	The subnets to place the ecs container instances in
- --region value		The AWS region to create the environment in
- --size value			Then number of container instances to create (default: 0)
- --instance-type value		The type of container instances to create

## up

Deploys the infrastructure for an ecso environment

All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.

````
ecso environment up [command options] ENVIRONMENT
````

#### Options
- --dry-run	If set, list pending changes, but do not execute the updates.

## rm

Removes an ecso environment

Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file

````
ecso environment rm [command options] ENVIRONMENT
````

#### Options
- --force	Required. Confirms the environment will be removed

## describe

Describes an ecso environment

````
ecso environment describe ENVIRONMENT
````
## down

Terminates an ecso environment

Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.

````
ecso environment down [command options] ENVIRONMENT
````

#### Options
- --force	Required. Confirms the environment will be stopped

# ecso service

Manage ecso services

````
ecso service command [command options] [arguments...]
````

#### Commands
- add		Adds a new service to the project
- up		Deploy a service
- down		terminates a service
- ls		List services
- ps		Show running tasks for a service
- events	List ECS events for a service
- logs		output service logs
- describe	Lists details of a deployed service


#### Options
   --help, -h	show help
   
## add

Adds a new service to the project

The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.

````
ecso service add [command options] SERVICE
````

#### Options
- --desired-count value	The desired number of service instances (default: 0)
- --route value		If set, the service will be registered with the load balancer at this route
- --port value		If set, the loadbalancer will bind to this port of the web container in this service (default: 0)

## up

Deploy a service

The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.

````
ecso service up [command options] SERVICE
````

#### Options
- --environment value	The name of the environment to deploy to [$ECSO_ENVIRONMENT]

## down

terminates a service

The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.

````
ecso service down [command options] SERVICE
````

#### Options
- --environment value	The environment to terminate the service from [$ECSO_ENVIRONMENT]

## ls

List services

````
ecso service ls [command options] [arguments...]
````

#### Options
- --environment value	Environment to query [$ECSO_ENVIRONMENT]

## ps

Show running tasks for a service

````
ecso service ps [command options] SERVICE
````

#### Options
- --environment value	The name of the environment [$ECSO_ENVIRONMENT]

## events

List ECS events for a service

````
ecso service events [command options] SERVICE
````

#### Options
- --environment value	The name of the environment [$ECSO_ENVIRONMENT]

## logs

output service logs

````
ecso service logs [command options] SERVICE
````

#### Options
- --environment value	The environment to terminate the service from [$ECSO_ENVIRONMENT]

## describe

Lists details of a deployed service

Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned

````
ecso service describe [command options] SERVICE
````

#### Options
- --environment value	The environment to query [$ECSO_ENVIRONMENT]

## env

Display the commands to set up the default environment for the ecso cli tool

````
main env [command options] ENVIRONMENT
````

#### Options
- --unset	If set, output shell commands to unset all ecso environment variables
