package ls

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type Options struct {
	Environment string
}

func New(env string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Environment: env,
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
		env      = ctx.Project.Environments[cmd.options.Environment]
		registry = ctx.Config.MustGetAWSClientRegistry(env.Region)
		ecsAPI   = registry.ECSAPI()
	)

	services, err := getServices(env, ecsAPI)

	if err != nil {
		return err
	}

	printServices(services)

	return nil
}

func (cmd *command) Prompt(ctx *ecso.CommandContext) error {
	return nil
}

func (cmd *command) Validate(ctx *ecso.CommandContext) error {
	opt := cmd.options

	if opt.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasEnvironment(opt.Environment) {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
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

func printServices(services []*ecs.Service) {
	headers := []string{"SERVICE", "TASK", "DESIRED", "RUNNING", "STATUS"}
	rows := make([]map[string]string, len(services))

	for i, service := range services {
		rows[i] = map[string]string{
			"SERVICE": *service.ServiceName,
			"TASK":    taskDefinitionName(*service.TaskDefinition),
			"DESIRED": fmt.Sprintf("%d", *service.DesiredCount),
			"RUNNING": fmt.Sprintf("%d", *service.RunningCount),
			"STATUS":  *service.Status,
		}
	}

	ui.PrintTable(os.Stdout, headers, rows...)
}

func taskDefinitionName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}

func serviceName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
