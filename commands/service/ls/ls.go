package ls

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/bernos/ecso/pkg/ecso"
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
	if err := validateOptions(cmd.options, ctx); err != nil {
		return err
	}

	env := ctx.Project.Environments[cmd.options.Environment]

	registry := ctx.Config.MustGetAWSClientRegistry(env.Region)
	// registry, err := ctx.Config.GetAWSClientRegistry(env.Region)

	// if err != nil {
	// 	return err
	// }

	ecsAPI := registry.ECSAPI()

	services, err := getServices(env, ecsAPI)

	if err != nil {
		return err
	}

	printServices(services, ctx.Config.Logger)

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

func printServices(services []*ecs.Service, log ecso.Logger) {
	var (
		pad             = 2
		longestName     = len("SERVICE")
		longestTaskName = len("TASK")
	)

	for _, svc := range services {
		name := *svc.ServiceName
		task := taskDefinitionName(*svc.TaskDefinition)

		if len(name) > longestName {
			longestName = len(name)
		}

		if len(task) > longestTaskName {
			longestTaskName = len(task)
		}
	}

	headerFormat := fmt.Sprintf(
		"\n%%-%ds%%-%ds%%-%ds%%-%ds%%s\n",
		longestName+pad,
		longestTaskName+pad,
		len("DESIRED")+pad,
		len("RUNNING")+pad)

	serviceFormat := fmt.Sprintf(
		"%%-%ds%%-%ds%%-%dd%%-%dd%%s\n",
		longestName+pad,
		longestTaskName+pad,
		len("DESIRED")+pad,
		len("RUNNING")+pad)

	log.Printf(headerFormat, "SERVICE", "TASK", "DESIRED", "RUNNING", "STATUS")

	for _, svc := range services {
		log.Printf(
			serviceFormat,
			*svc.ServiceName,
			taskDefinitionName(*svc.TaskDefinition),
			*svc.DesiredCount,
			*svc.RunningCount,
			*svc.Status)
	}
}

func validateOptions(opt *Options, ctx *ecso.CommandContext) error {
	if opt.Environment == "" {
		return fmt.Errorf("Environment is required")
	}

	if !ctx.Project.HasEnvironment(opt.Environment) {
		return fmt.Errorf("Environment '%s' not found", opt.Environment)
	}

	return nil
}

func taskDefinitionName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}

func serviceName(arn string) string {
	tokens := strings.Split(arn, "/")
	return tokens[len(tokens)-1]
}
