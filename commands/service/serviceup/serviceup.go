package serviceup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/services"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type Options struct {
	Name        string
	Environment string
}

type command struct {
	options *Options
}

func New(name, environment string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name:        name,
		Environment: environment,
	}

	for _, option := range options {
		option(o)
	}

	return &command{
		options: o,
	}
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	var (
		cfg         = ctx.Config
		log         = cfg.Logger
		project     = ctx.Project
		environment = ctx.Project.Environments[cmd.options.Environment]
		service     = project.Services[cmd.options.Name]
	)

	log.BannerBlue(
		"Deploying service '%s' to the '%s' environment",
		service.Name,
		environment.Name)

	// Setup env vars
	if err := setEnv(project, environment, service); err != nil {
		return err
	}

	// Deploy cfn stack
	if err := deployStack(ctx, environment, service); err != nil {
		return err
	}

	// Deploy the ecs service
	if err := deployService(ctx, environment, service); err != nil {
		return err
	}

	log.BannerGreen(
		"Deployed service '%s' to the '%s' environment",
		service.Name,
		environment.Name)

	return logOutputs(ctx, environment, service)
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	err := util.AnyError(
		ui.ValidateRequired("Name")(opt.Name),
		ui.ValidateRequired("Environment")(opt.Environment))

	if err != nil {
		return err
	}

	if _, ok := ctx.Project.Services[opt.Name]; !ok {
		return fmt.Errorf("Service '%s' not found", opt.Name)
	}

	if _, ok := ctx.Project.Environments[opt.Environment]; !ok {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
	}

	return nil
}

func logOutputs(ctx *ecso.CommandContext, env *ecso.Environment, service *ecso.Service) error {
	var (
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		cfn      = registry.CloudFormationService(ctx.Config.Logger.PrefixPrintf("  "))
	)

	outputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return err
	}

	if service.Route != "" {
		ctx.Config.Logger.Dt(
			"Service URL",
			fmt.Sprintf("%s%s", outputs["LoadBalancerUrl"], service.Route))
	}

	ctx.Config.Logger.Dt(
		"Service console",
		util.ServiceConsoleURL(service.GetECSServiceName(), env.GetClusterName(), env.Region))

	ctx.Config.Logger.Printf("\n")

	return nil
}

func setEnv(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
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

func deployService(ctx *ecso.CommandContext, env *ecso.Environment, service *ecso.Service) error {
	var (
		cfg = ctx.Config
		log = cfg.Logger

		cluster        = env.GetClusterName()
		stackName      = service.GetCloudFormationStackName(env)
		taskName       = service.GetECSTaskDefinitionName(env)
		ecsServiceName = service.GetECSServiceName()

		registry   = cfg.MustGetAWSClientRegistry(env.Region)
		ecsClient  = registry.ECSAPI()
		cfnService = registry.CloudFormationService(log.PrefixPrintf("  "))
		ecsService = registry.ECSService(log.PrefixPrintf("  "))
	)

	serviceStackOutputs, err := cfnService.GetStackOutputs(stackName)

	if err != nil {
		return err
	}

	// TODO: fully qualify the path to the service compose file
	// taskDefinition, err := ConvertToTaskDefinition(taskName, service.ComposeFile)
	log.Printf("\n")
	log.Infof("Converting '%s' to task definition...", service.ComposeFile)

	taskDefinition, err := service.GetECSTaskDefinition(env)

	log.Printf("\n")
	log.Infof("Registering ECS task definition '%s'...", taskName)

	if err != nil {
		return err
	}

	for _, container := range taskDefinition.ContainerDefinitions {
		container.SetLogConfiguration(&ecs.LogConfiguration{
			LogDriver: aws.String(ecs.LogDriverAwslogs),
			Options: map[string]*string{
				"awslogs-region": aws.String(env.Region),
				"awslogs-group":  aws.String(serviceStackOutputs["CloudWatchLogsGroup"]),
			},
		})
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
		return err
	}

	log.Printf(
		"  Registered ECS task definition %s:%d\n\n",
		*resp.TaskDefinition.Family,
		*resp.TaskDefinition.Revision)

	services, err := ecsClient.DescribeServices(&ecs.DescribeServicesInput{
		Services: []*string{
			aws.String(ecsServiceName),
		},
		Cluster: aws.String(cluster),
	})

	if err != nil {
		return err
	}

	log.Infof("Deploying ECS service '%s'", ecsServiceName)

	isCreate := true

	for _, s := range services.Services {
		if *s.Status != "INACTIVE" {
			isCreate = false
		}
	}

	if isCreate {
		log.Printf("  Creating new ecs service...\n")

		input := &ecs.CreateServiceInput{
			DesiredCount:   aws.Int64(int64(service.DesiredCount)),
			ServiceName:    aws.String(ecsServiceName),
			TaskDefinition: resp.TaskDefinition.TaskDefinitionArn,
			Cluster:        aws.String(cluster),
			DeploymentConfiguration: &ecs.DeploymentConfiguration{
				MaximumPercent:        aws.Int64(200),
				MinimumHealthyPercent: aws.Int64(100),
			},
		}

		if len(service.Route) > 0 {
			input.LoadBalancers = []*ecs.LoadBalancer{
				{
					ContainerName:  aws.String("web"),
					ContainerPort:  aws.Int64(int64(service.Port)),
					TargetGroupArn: aws.String(serviceStackOutputs["TargetGroup"]),
				},
			}

			input.Role = aws.String(serviceStackOutputs["ServiceRole"])
		}

		_, err := ecsClient.CreateService(input)

		if err != nil {
			return err
		}

		log.Printf("  Create successful\n")
	} else {
		log.Printf("  Updating existing ecs service...\n")

		_, err := ecsClient.UpdateService(&ecs.UpdateServiceInput{
			DesiredCount:   aws.Int64(int64(service.DesiredCount)),
			Service:        aws.String(ecsServiceName),
			TaskDefinition: resp.TaskDefinition.TaskDefinitionArn,
			Cluster:        aws.String(cluster),
			DeploymentConfiguration: &ecs.DeploymentConfiguration{
				MaximumPercent:        aws.Int64(200),
				MinimumHealthyPercent: aws.Int64(100),
			},
		})

		if err != nil {
			return err
		}

		log.Printf("  Update successful\n")
	}

	log.Printf("\n")
	log.Infof("Waiting for service to become stable...")

	cancel := ecsService.LogServiceEvents(ecsServiceName, env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
		if err == nil && e != nil {
			log.Printf("  %s %s\n", *e.CreatedAt, *e.Message)
		}
	})

	defer cancel()

	if err := ecsClient.WaitUntilServicesStable(&ecs.DescribeServicesInput{
		Services: []*string{
			aws.String(ecsServiceName),
		},
		Cluster: aws.String(cluster),
	}); err != nil {
		return err
	}

	return nil
}

func deployStack(ctx *ecso.CommandContext, env *ecso.Environment, service *ecso.Service) error {
	var (
		cfg        = ctx.Config
		log        = cfg.Logger
		project    = ctx.Project
		bucket     = env.CloudFormationBucket
		stackName  = service.GetCloudFormationStackName(env)
		prefix     = service.GetCloudFormationBucketPrefix(env)
		registry   = cfg.MustGetAWSClientRegistry(env.Region)
		template   = service.GetCloudFormationTemplateFile()
		cfnService = registry.CloudFormationService(log.PrefixPrintf("  "))
	)

	params, err := getCloudFormationParameters(cfnService, project, env, service)

	if err != nil {
		return err
	}

	tags := getCloudFormationTags(project, env, service)

	log.Infof("Deploying service cloudformartion stack '%s'...", stackName)

	result, err := cfnService.PackageAndDeploy(
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

func getTemplateDir(serviceName string) (string, error) {
	wd, err := ecso.GetCurrentProjectDir()

	if err != nil {
		return wd, err
	}

	return filepath.Join(wd, ".ecso", "services", serviceName), nil
}

func getCloudFormationParameters(cfnService services.CloudFormationService, project *ecso.Project, env *ecso.Environment, service *ecso.Service) (map[string]string, error) {

	outputs, err := cfnService.GetStackOutputs(fmt.Sprintf("%s-%s", project.Name, env.Name))

	if err != nil {
		return nil, err
	}

	var params map[string]string

	if len(service.Route) == 0 {
		params = make(map[string]string)
	} else {
		params = map[string]string{
			"VPC":           outputs["VPC"],
			"Listener":      outputs["Listener"],
			"Path":          service.Route,
			"RoutePriority": strconv.Itoa(service.RoutePriority),
		}
	}

	for k, v := range service.Environments[env.Name].CloudFormationParameters {
		params[k] = v
	}

	return params, nil
}

func getCloudFormationTags(project *ecso.Project, env *ecso.Environment, service *ecso.Service) map[string]string {
	tags := map[string]string{
		"project":     project.Name,
		"environment": env.Name,
	}

	for k, v := range service.Tags {
		tags[k] = v
	}

	return tags
}
