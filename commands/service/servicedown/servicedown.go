package servicedown

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/services"
)

type Options struct {
	Name        string
	Environment string
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

type command struct {
	options *Options
}

func (cmd *command) Execute(ctx *ecso.CommandContext) error {
	var (
		service    = ctx.Project.Services[cmd.options.Name]
		env        = ctx.Project.Environments[cmd.options.Environment]
		log        = ctx.Config.Logger
		registry   = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsAPI     = registry.ECSAPI()
		ecsService = registry.ECSService(log.PrefixPrintf("  "))
		cfnService = registry.CloudFormationService(log.PrefixPrintf("  "))
	)

	log.BannerBlue(
		"Terminating the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	exists, err := ecsServiceExists(service, env, ecsAPI)

	if err != nil {
		return err
	}

	if exists {
		log.Infof("Stopping ECS service '%s'", service.GetECSServiceName())

		if err := stopECSService(ecsService, ecsAPI, service, env, log.PrefixPrintf("  ")); err != nil {
			return err
		}

		log.Printf("\n")
		log.Infof("Deleting ECS service '%s'", service.GetECSServiceName())

		if _, err := ecsAPI.DeleteService(&ecs.DeleteServiceInput{
			Cluster: aws.String(env.GetClusterName()),
			Service: aws.String(service.GetECSServiceName()),
		}); err != nil {
			return err
		}
	} else {
		log.Infof("ECS service '%s' doesn't exists, nothing to clean up", service.GetECSServiceName())
	}

	log.Printf("\n")
	log.Infof("Deleting cloud formation stack '%s'", service.GetCloudFormationStackName(env))

	if err := deleteCloudFormationStack(cfnService, service, env, log.PrefixPrintf("  ")); err != nil {
		return err
	}

	log.BannerGreen(
		"Successfully terminated the '%s' service in the '%s' environment",
		service.Name,
		env.Name)

	// delete the cfn stack
	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if opt.Name == "" {
		return fmt.Errorf("Name is required")
	}

	if opt.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasService(opt.Name) {
		return fmt.Errorf("No service named '%s' was found", opt.Name)
	}

	if !ctx.Project.HasEnvironment(opt.Environment) {
		return fmt.Errorf("No environment named '%s' was found", opt.Environment)
	}

	return nil
}

func deleteCloudFormationStack(cfnService services.CloudFormationService, service *ecso.Service, env *ecso.Environment, log func(string, ...interface{})) error {
	stackName := service.GetCloudFormationStackName(env)
	exists, err := cfnService.StackExists(stackName)

	if err != nil {
		return nil
	}

	if !exists {
		log("Stack '%s' does not exist\n", stackName)
		return nil
	}

	return cfnService.DeleteStack(stackName)
}

func ecsServiceExists(service *ecso.Service, env *ecso.Environment, ecsAPI ecsiface.ECSAPI) (bool, error) {
	resp, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: aws.String(env.GetClusterName()),
		Services: []*string{
			aws.String(service.GetECSServiceName()),
		},
	})

	if err != nil {
		return false, err
	}

	return len(resp.Services) > 0, nil
}

func stopECSService(ecsService services.ECSService, ecsAPI ecsiface.ECSAPI, service *ecso.Service, env *ecso.Environment, log func(string, ...interface{})) error {

	describeServiceInput := &ecs.DescribeServicesInput{
		Cluster: aws.String(env.GetClusterName()),
		Services: []*string{
			aws.String(service.GetECSServiceName()),
		},
	}

	// First check if the service is running
	description, err := ecsAPI.DescribeServices(describeServiceInput)

	if err != nil {
		return err
	}

	// Nothing to do
	if len(description.Services) == 0 {
		log("No service named '%s' was found in the cluster '%s'", service.GetECSServiceName(), env.GetClusterName())
		return nil
	}

	if len(description.Services) > 1 {
		return fmt.Errorf("Found more than one ecs service named '%s'", service.GetECSServiceName())
	}

	status := *description.Services[0].Status

	if status == "ACTIVE" {
		log("Setting desired count to 0...\n")

		_, err = ecsAPI.UpdateService(&ecs.UpdateServiceInput{
			Cluster:      aws.String(env.GetClusterName()),
			Service:      aws.String(service.GetECSServiceName()),
			DesiredCount: aws.Int64(0),
		})

		if err != nil {
			return err
		}

		log("Waiting for tasks to drain, and service to become stable...\n")

		cancel := ecsService.LogServiceEvents(service.GetECSServiceName(), env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
			if err == nil && e != nil {
				log("  %s %s\n", *e.CreatedAt, *e.Message)
			}
		})

		defer cancel()

		if err := ecsAPI.WaitUntilServicesStable(describeServiceInput); err != nil {
			return err
		}

		log("Deleting service...\n")

		_, err = ecsAPI.DeleteService(&ecs.DeleteServiceInput{
			Cluster: aws.String(env.GetClusterName()),
			Service: aws.String(service.GetECSServiceName()),
		})

		if err != nil {
			return err
		}

		log("Waiting for service to become inactive...\n")

		if err := ecsAPI.WaitUntilServicesInactive(describeServiceInput); err != nil {
			return err
		}
	} else if status == "DRAINING" {

		log("Waiting for service to become inactive...\n")

		if err := ecsAPI.WaitUntilServicesInactive(describeServiceInput); err != nil {
			return err
		}
	} else if status == "INACTIVE" {
		log("Service is already inactive\n")
	}

	return nil
}
