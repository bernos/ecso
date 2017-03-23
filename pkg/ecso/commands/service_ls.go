package commands

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
	"gopkg.in/urfave/cli.v1"
)

const (
	ServiceLsEnvironmentOption = "environment"
)

func NewServiceLsCommand(env string, log log.Logger, registryFactory awsregistry.RegistryFactory) ecso.Command {
	return &serviceLsCommand{
		environment:     env,
		log:             log,
		registryFactory: registryFactory,
	}
}

type serviceLsCommand struct {
	environment     string
	log             log.Logger
	registryFactory awsregistry.RegistryFactory
}

func (cmd *serviceLsCommand) UnmarshalCliContext(ctx *cli.Context) error {
	return nil
}

func (cmd *serviceLsCommand) Execute(ctx *ecso.CommandContext) error {
	env := ctx.Project.Environments[cmd.environment]

	reg, err := cmd.registryFactory.ForRegion(env.Region)

	if err != nil {
		return err
	}

	services, err := getServices(env, reg.ECSAPI())

	if err != nil {
		return err
	}

	printServices(ctx.Project, env, services, cmd.log)

	return nil
}

func (cmd *serviceLsCommand) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *serviceLsCommand) Validate(ctx *ecso.CommandContext) error {
	if cmd.environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasEnvironment(cmd.environment) {
		return fmt.Errorf("Environment '%s' not found", cmd.environment)
	}

	return nil
}

func getServices(env *ecso.Environment, ecsAPI ecsiface.ECSAPI) ([]*ecs.Service, error) {
	var (
		count    = 0
		batches  = make([][]*string, 0)
		services = make([]*ecs.Service, 0)
	)

	params := &ecs.ListServicesInput{
		Cluster: aws.String(env.GetClusterName()),
	}

	if err := ecsAPI.ListServicesPages(params, func(o *ecs.ListServicesOutput, last bool) bool {
		if count%10 == 0 {
			batches = append(batches, make([]*string, 0))
		}

		for i, _ := range o.ServiceArns {
			batch := append(batches[len(batches)-1], o.ServiceArns[i])
			batches[len(batches)-1] = batch
			count = count + 1
		}

		return !last
	}); err != nil {
		return services, err
	}

	for _, batch := range batches {
		if len(batch) == 0 {
			continue
		}

		desc, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
			Services: batch,
			Cluster:  aws.String(env.GetClusterName()),
		})

		if err != nil {
			return services, err
		}

		for _, svc := range desc.Services {
			services = append(services, svc)
		}
	}

	return services, nil
}

func printServices(project *ecso.Project, env *ecso.Environment, services []*ecs.Service, log log.Logger) {
	headers := []string{"SERVICE", "ECS SERVICE", "TASK", "DESIRED", "RUNNING", "STATUS"}
	rows := make([]map[string]string, len(services))

	for i, service := range services {
		rows[i] = map[string]string{
			"SERVICE":     localServiceName(*service.ServiceName, env, project),
			"ECS SERVICE": *service.ServiceName,
			"TASK":        taskDefinitionName(*service.TaskDefinition),
			"DESIRED":     fmt.Sprintf("%d", *service.DesiredCount),
			"RUNNING":     fmt.Sprintf("%d", *service.RunningCount),
			"STATUS":      *service.Status,
		}
	}

	ui.PrintTable(log, headers, rows...)
}

func localServiceName(ecsServiceName string, env *ecso.Environment, project *ecso.Project) string {
	for _, s := range project.Services {
		if strings.HasPrefix(ecsServiceName, s.GetECSTaskDefinitionName(env)+"-Service") {
			return s.Name
		}
	}

	return ""
}

func taskDefinitionName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}

func serviceName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
