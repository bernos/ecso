# ecso init

```
NAME:
   ecso init - Initialise a new ecso project

USAGE:
   ecso init [PROJECT]

DESCRIPTION:
   Creates a new ecso project configuration file at .ecso/project.json. The initial project contains no environments or services. The project configuration file can be safely endited by hand, but it is usually easier to user the ecso cli tool to add new services and environments to the project.
```

# ecso environment add

```
NAME:
   ecso environment add - Add a new environment to the project

USAGE:
   ecso environment add [command options] [ENVIRONMENT]

OPTIONS:
   --vpc value			The vpc to create the environment in
   --alb-subnets value		The subnets to place the application load balancer in
   --instance-subnets value	The subnets to place the ecs container instances in
   --region value		The AWS region to create the environment in
   --size value			Then number of container instances to create (default: 0)
   --instance-type value	The type of container instances to create
   
```

# ecso environment up

```
NAME:
   ecso environment up - Deploys the infrastructure for an ecso environment

USAGE:
   ecso environment up [command options] ENVIRONMENT

DESCRIPTION:
   All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.

OPTIONS:
   --dry-run	If set, list pending changes, but do not execute the updates.
   
```

# ecso environment describe

```
NAME:
   ecso environment describe - Describes an ecso environment

USAGE:
   ecso environment describe ENVIRONMENT
```

# ecso environment down

```
NAME:
   ecso environment down - Terminates an ecso environment

USAGE:
   ecso environment down [command options] ENVIRONMENT

DESCRIPTION:
   Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.

OPTIONS:
   --force	Required. Confirms the environment will be stopped
   
```

# ecso environment rm

```
NAME:
   ecso environment rm - Removes an ecso environment

USAGE:
   ecso environment rm [command options] ENVIRONMENT

DESCRIPTION:
   Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file

OPTIONS:
   --force	Required. Confirms the environment will be removed
   
```

# ecso env

```
NAME:
   ecso env - Display the commands to set up the default environment for the ecso cli tool

USAGE:
   ecso env [command options] ENVIRONMENT

OPTIONS:
   --unset	If set, output shell commands to unset all ecso environment variables
   
```

# ecso service add

```
NAME:
   ecso service add - Adds a new service to the project

USAGE:
   ecso service add [command options] SERVICE

DESCRIPTION:
   The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.

OPTIONS:
   --desired-count value	The desired number of service instances (default: 0)
   --route value		If set, the service will be registered with the load balancer at this route
   --port value			If set, the loadbalancer will bind to this port of the web container in this service (default: 0)
   
```

# ecso service up

```
NAME:
   ecso service up - Deploy a service

USAGE:
   ecso service up [command options] SERVICE

DESCRIPTION:
   The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.

OPTIONS:
   --environment value	The name of the environment to deploy to [$ECSO_ENVIRONMENT]
   
```

# ecso service down

```
NAME:
   ecso service down - terminates a service

USAGE:
   ecso service down [command options] SERVICE

DESCRIPTION:
   The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.

OPTIONS:
   --environment value	The environment to terminate the service from [$ECSO_ENVIRONMENT]
   
```

# ecso service ls

```
NAME:
   ecso service ls - List services

USAGE:
   ecso service ls [command options] [arguments...]

OPTIONS:
   --environment value	Environment to query [$ECSO_ENVIRONMENT]
   
```

# ecso service ps

```
NAME:
   ecso service ps - Show running tasks for a service

USAGE:
   ecso service ps [command options] SERVICE

OPTIONS:
   --environment value	The name of the environment [$ECSO_ENVIRONMENT]
   
```

# ecso service logs

```
NAME:
   ecso service logs - output service logs

USAGE:
   ecso service logs [command options] SERVICE

OPTIONS:
   --environment value	The environment to terminate the service from [$ECSO_ENVIRONMENT]
   
```

# ecso service describe

```
NAME:
   ecso service describe - Lists details of a deployed service

USAGE:
   ecso service describe [command options] SERVICE

DESCRIPTION:
   Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned

OPTIONS:
   --environment value	The environment to query [$ECSO_ENVIRONMENT]
   
```

