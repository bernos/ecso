package api

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/services"
	"github.com/bernos/ecso/pkg/ecso/util"
)

func (api *api) ServiceUp(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	// set env vars so that they are available when converting the docker
	// compose file to a task definition
	if err := api.setEnv(project, env, service); err != nil {
		return err
	}

	// register task
	taskDefinition, err := api.registerECSTaskDefinition(reg, project, env, service)

	if err != nil {
		return err
	}

	// deploy the service cfn stack
	if err := api.deployServiceStack(reg, project, env, service, taskDefinition); err != nil {
		return err
	}

	return nil
}

func (api *api) setEnv(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	if err := util.AnyError(
		os.Setenv("ECSO_ENVIRONMENT", env.Name),
		os.Setenv("ECSO_AWS_REGION", env.Region),
		os.Setenv("ECSO_CLUSTER_NAME", env.GetClusterName())); err != nil {
		return err
	}

	// set any env vars from the service configuration for the current environment
	for k, v := range service.Environments[env.Name].Env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (api *api) deployServiceStack(reg *ecso.AWSClientRegistry, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition) error {
	var (
		cfg       = api.cfg
		log       = cfg.Logger()
		bucket    = env.CloudFormationBucket
		stackName = service.GetCloudFormationStackName(env)
		prefix    = service.GetCloudFormationBucketPrefix(env)
		template  = service.GetCloudFormationTemplateFile()
		cfn       = reg.CloudFormationService(log.PrefixPrintf("  "))
	)

	params, err := getServiceStackParameters(cfn, project, env, service, taskDefinition)

	if err != nil {
		return err
	}

	tags := getServiceStackTags(project, env, service)

	log.Infof("Deploying service cloudformation stack '%s'...", stackName)

	result, err := cfn.PackageAndDeploy(
		stackName,
		template,
		bucket,
		prefix,
		tags,
		params,
		false)

	if err != nil {
		return err
	}

	if !result.DidRequireUpdating {
		log.Printf("  No updates were required to Cloud Formation stack '%s'\n", result.StackID)
	}

	return nil
}

func getServiceStackParameters(cfn services.CloudFormationService, project *ecso.Project, env *ecso.Environment, service *ecso.Service, taskDefinition *ecs.TaskDefinition) (map[string]string, error) {

	outputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	params := map[string]string{
		"Cluster":        outputs["Cluster"],
		"AlertsTopic":    outputs["AlertsTopic"],
		"DesiredCount":   fmt.Sprintf("%d", service.DesiredCount),
		"TaskDefinition": *taskDefinition.TaskDefinitionArn,
	}

	if len(service.Route) > 0 {
		params["VPC"] = outputs["VPC"]
		params["Listener"] = outputs["Listener"]
		params["Path"] = service.Route
		params["Port"] = fmt.Sprintf("%d", service.Port)
		params["RoutePriority"] = strconv.Itoa(service.RoutePriority)
	}

	for k, v := range service.Environments[env.Name].CloudFormationParameters {
		params[k] = v
	}

	return params, nil
}

func getServiceStackTags(project *ecso.Project, env *ecso.Environment, service *ecso.Service) map[string]string {
	tags := map[string]string{
		"project":     project.Name,
		"environment": env.Name,
	}

	for k, v := range service.Tags {
		tags[k] = v
	}

	return tags
}

func (api *api) registerECSTaskDefinition(reg *ecso.AWSClientRegistry, project *ecso.Project, env *ecso.Environment, service *ecso.Service) (*ecs.TaskDefinition, error) {
	var (
		cfg       = api.cfg
		log       = cfg.Logger()
		taskName  = service.GetECSTaskDefinitionName(env)
		ecsClient = reg.ECSAPI()
	)

	// TODO: fully qualify the path to the service compose file
	// taskDefinition, err := ConvertToTaskDefinition(taskName, service.ComposeFile)
	log.Printf("\n")
	log.Infof("Converting '%s' to task definition...", service.ComposeFile)

	taskDefinition, err := service.GetECSTaskDefinition(env)

	log.Printf("\n")
	log.Infof("Registering ECS task definition '%s'...", taskName)

	if err != nil {
		return nil, err
	}

	for _, container := range taskDefinition.ContainerDefinitions {
		container.SetLogConfiguration(&ecs.LogConfiguration{
			LogDriver: aws.String(ecs.LogDriverAwslogs),
			Options: map[string]*string{
				"awslogs-region": aws.String(env.Region),
				"awslogs-group":  aws.String(service.GetCloudWatchLogGroupName(env)),
			},
		})

		// TODO probably don't automatically add service discovery env config
		//      if people want to use service discover they can just add the
		//      required env vars to the services docker-compose file. This is
		//      less magic and more flexible
		for _, p := range container.PortMappings {
			container.Environment = append(container.Environment, &ecs.KeyValuePair{
				Name:  aws.String(fmt.Sprintf("SERVICE_%d_NAME", *p.ContainerPort)),
				Value: aws.String(fmt.Sprintf("%s.%s", service.Name, env.GetClusterName())),
			})
		}
	}

	resp, err := ecsClient.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: taskDefinition.ContainerDefinitions,
		Family:               taskDefinition.Family,
		NetworkMode:          taskDefinition.NetworkMode,
		PlacementConstraints: taskDefinition.PlacementConstraints,
		TaskRoleArn:          taskDefinition.TaskRoleArn,
		Volumes:              taskDefinition.Volumes,
	})

	if err != nil {
		return nil, err
	}

	log.Printf(
		"  Registered ECS task definition %s:%d\n\n",
		*resp.TaskDefinition.Family,
		*resp.TaskDefinition.Revision)

	return resp.TaskDefinition, nil
}
