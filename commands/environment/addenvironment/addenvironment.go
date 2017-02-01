package addenvironment

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
)

type Options struct {
	Name                 string
	CloudFormationBucket string
	VPCID                string
	ALBSubnets           string
	InstanceSubnets      string
	Region               string
	Account              string
	InstanceType         string
	Size                 int
	DNSZone              string
}

func New(environmentName string, options ...func(*Options)) ecso.Command {
	o := &Options{
		Name: environmentName,
	}

	for _, option := range options {
		option(o)
	}

	return &cmd{
		options: o,
	}
}

type cmd struct {
	options *Options
}

func (c *cmd) Execute(ctx *ecso.CommandContext) error {
	var (
		log     = ctx.Config.Logger
		project = ctx.Project
	)

	if project.HasEnvironment(c.options.Name) {
		return fmt.Errorf("An environment named '%s' already exists for this project.", c.options.Name)
	}

	project.AddEnvironment(&ecso.Environment{
		Name:                 c.options.Name,
		Region:               c.options.Region,
		CloudFormationBucket: c.options.CloudFormationBucket,
		CloudFormationParameters: map[string]string{
			"VPC":             c.options.VPCID,
			"InstanceSubnets": c.options.InstanceSubnets,
			"ALBSubnets":      c.options.ALBSubnets,
			"InstanceType":    c.options.InstanceType,
			"DNSZone":         c.options.DNSZone,
			"ClusterSize":     fmt.Sprintf("%d", c.options.Size),
		},
		CloudFormationTags: map[string]string{
			"environment": c.options.Name,
			"project":     project.Name,
		},
	})

	if err := project.Save(); err != nil {
		return err
	}

	log.BannerGreen("Successfully added environment '%s' to the project", c.options.Name)

	return nil
}

func (c *cmd) Validate(ctx *ecso.CommandContext) error {
	return nil
}
