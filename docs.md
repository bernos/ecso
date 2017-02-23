# ECSO 

#### Table of contents

- [init](#init)
- [environment](#environment)
  * [add](#environment-add)
  * [up](#environment-up)
  * [rm](#environment-rm)
  * [describe](#environment-describe)
  * [down](#environment-down)
- [service](#service)
  * [add](#service-add)
  * [up](#service-up)
  * [down](#service-down)
  * [ls](#service-ls)
  * [ps](#service-ps)
  * [events](#service-events)
  * [logs](#service-logs)
  * [describe](#service-describe)
- [env](#env)
- [help](#help)

<a id="init"></a>
# init

Initialise a new ecso project

````
ecso init [PROJECT]
````

<a id="environment"></a>
# environment

Manage ecso environments

````
ecso environment <command> [arguments...]
````

#### Commands
| Name  | Description |
|:---   |:---         |
| [add](#environment-add) | Add a new environment to the project | 
| [up](#environment-up) | Deploys the infrastructure for an ecso environment | 
| [rm](#environment-rm) | Removes an ecso environment | 
| [describe](#environment-describe) | Describes an ecso environment | 
| [down](#environment-down) | Terminates an ecso environment | 


<a id="environment-add"></a>
## add

Add a new environment to the project

````
ecso environment add [command options] [ENVIRONMENT]
````

#### Options
| option | usage |
|:---    |:---   |
| --vpc | The vpc to create the environment in |
| --alb-subnets | The subnets to place the application load balancer in |
| --instance-subnets | The subnets to place the ecs container instances in |
| --region | The AWS region to create the environment in |
| --size | Then number of container instances to create |
| --instance-type | The type of container instances to create |

<a id="environment-up"></a>
## up

Deploys the infrastructure for an ecso environment

All ecso environment infrastructure deployments are managed by CloudFormation. CloudFormation templates for environment infrastructure are stored at .ecso/infrastructure/templates, and are created the first time that `ecso environment up` is run. These templates can be safely edited by hand.

````
ecso environment up [command options] ENVIRONMENT
````

#### Options
| option | usage |
|:---    |:---   |
| --dry-run | If set, list pending changes, but do not execute the updates. |

<a id="environment-rm"></a>
## rm

Removes an ecso environment

Terminates an environment if it is running, and also deletes the environment configuration from the .ecso/project.json file

````
ecso environment rm [command options] ENVIRONMENT
````

#### Options
| option | usage |
|:---    |:---   |
| --force | Required. Confirms the environment will be removed |

<a id="environment-describe"></a>
## describe

Describes an ecso environment

````
ecso environment describe ENVIRONMENT
````

<a id="environment-down"></a>
## down

Terminates an ecso environment

Any services running in the environment will be terminated first. See the description of 'ecso service down' for details. Once all running services have been terminated, the environment Cloud Formation stack will be deleted, and any DNS entries removed.

````
ecso environment down [command options] ENVIRONMENT
````

#### Options
| option | usage |
|:---    |:---   |
| --force | Required. Confirms the environment will be stopped |

<a id="service"></a>
# service

Manage ecso services

````
ecso service <command> [arguments...]
````

#### Commands
| Name  | Description |
|:---   |:---         |
| [add](#service-add) | Adds a new service to the project | 
| [up](#service-up) | Deploy a service | 
| [down](#service-down) | terminates a service | 
| [ls](#service-ls) | List services | 
| [ps](#service-ps) | Show running tasks for a service | 
| [events](#service-events) | List ECS events for a service | 
| [logs](#service-logs) | output service logs | 
| [describe](#service-describe) | Lists details of a deployed service | 


<a id="service-add"></a>
## add

Adds a new service to the project

The .ecso/project.json file will be updated with configuration settings for the new service. CloudFormation templates for the service and supporting resources are created in the .ecso/services/SERVICE dir, and can be safely edited by hand. An initial docker compose file will be created at ./services/SERVICE/docker-compose.yaml.

````
ecso service add [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --desired-count | The desired number of service instances |
| --route | If set, the service will be registered with the load balancer at this route |
| --port | If set, the loadbalancer will bind to this port of the web container in this service |

<a id="service-up"></a>
## up

Deploy a service

The service's docker-compose file will be transformed into an ECS task definition, and registered with ECS. The service CloudFormation template will be deployed. Service deployment policies and constraints can be set in the service CloudFormation templates. By default a rolling deployment is performed, with the number of services running at any time equal to at least the desired service count, and at most 200% of the desired service count.

````
ecso service up [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The name of the environment to deploy to |

<a id="service-down"></a>
## down

terminates a service

The service will be scaled down, then deleted. The service's CloudFormation stack will be deleted, and any DNS records removed.

````
ecso service down [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The environment to terminate the service from |

<a id="service-ls"></a>
## ls

List services

````
ecso service ls [command options] [arguments...]
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | Environment to query |

<a id="service-ps"></a>
## ps

Show running tasks for a service

````
ecso service ps [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The name of the environment |

<a id="service-events"></a>
## events

List ECS events for a service

````
ecso service events [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The name of the environment |

<a id="service-logs"></a>
## logs

output service logs

````
ecso service logs [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The environment to terminate the service from |

<a id="service-describe"></a>
## describe

Lists details of a deployed service

Returns detailed information about a deployed service. If the service has not been deployed to the environment an error will be returned

````
ecso service describe [command options] SERVICE
````

#### Options
| option | usage |
|:---    |:---   |
| --environment | The environment to query |

<a id="env"></a>
# env

Display the commands to set up the default environment for the ecso cli tool

````
ecso env [command options] ENVIRONMENT
````

#### Options
| option | usage |
|:---    |:---   |
| --unset | If set, output shell commands to unset all ecso environment variables |

<a id="help"></a>
# help

Shows a list of commands or help for one command

````
ecso help [command]
````
