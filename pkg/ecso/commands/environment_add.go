package commands

import (
	"fmt"
	"io"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/api"
)

type EnvironmentAddCommand struct {
	*EnvironmentCommand

	vpcID           string
	albSubnets      string
	instanceSubnets string
	region          string
	instanceType    string
	size            int
	keyPair         string
	dnsZone         string
	datadogAPIKey   string
}

func (c *EnvironmentAddCommand) WithDatadogAPIKey(apiKey string) *EnvironmentAddCommand {
	c.datadogAPIKey = apiKey
	return c
}

func (c *EnvironmentAddCommand) WithDNSZone(zone string) *EnvironmentAddCommand {
	c.dnsZone = zone
	return c
}

func (c *EnvironmentAddCommand) WithKeyPair(keyPair string) *EnvironmentAddCommand {
	c.keyPair = keyPair
	return c
}

func (c *EnvironmentAddCommand) WithSize(size int) *EnvironmentAddCommand {
	c.size = size
	return c
}

func (c *EnvironmentAddCommand) WithInstanceType(instanceType string) *EnvironmentAddCommand {
	c.instanceType = instanceType
	return c
}

func (c *EnvironmentAddCommand) WithRegion(region string) *EnvironmentAddCommand {
	c.region = region
	return c
}

func (c *EnvironmentAddCommand) WithInstanceSubnets(subnets string) *EnvironmentAddCommand {
	c.instanceSubnets = subnets
	return c
}

func (c *EnvironmentAddCommand) WithALBSubnets(subnets string) *EnvironmentAddCommand {
	c.albSubnets = subnets
	return c
}

func (c *EnvironmentAddCommand) WithVPCID(vpcID string) *EnvironmentAddCommand {
	c.vpcID = vpcID
	return c
}

func NewEnvironmentAddCommand(environmentName string, environmentAPI api.EnvironmentAPI) *EnvironmentAddCommand {
	return &EnvironmentAddCommand{
		EnvironmentCommand: &EnvironmentCommand{
			environmentName: environmentName,
			environmentAPI:  environmentAPI,
		},
	}
}

func (c *EnvironmentAddCommand) Execute(ctx *ecso.CommandContext, r io.Reader, w io.Writer) error {
	if ctx.Project.HasEnvironment(c.environmentName) {
		return ecso.NewEnvironmentExistsError(c.environmentName)
	}

	ctx.Project.AddEnvironment(&ecso.Environment{
		Name:   c.environmentName,
		Region: c.region,
		CloudFormationParameters: map[string]string{
			"VPC":             c.vpcID,
			"InstanceSubnets": c.instanceSubnets,
			"ALBSubnets":      c.albSubnets,
			"InstanceType":    c.instanceType,
			"DNSZone":         c.dnsZone,
			"ClusterSize":     fmt.Sprintf("%d", c.size),
			"DataDogAPIKey":   c.datadogAPIKey,
			"KeyPair":         c.keyPair,
		},
		CloudFormationTags: map[string]string{
			"environment": c.environmentName,
			"project":     ctx.Project.Name,
		},
	})

	return ctx.Project.Save()
}

func (c *EnvironmentAddCommand) Validate(ctx *ecso.CommandContext) error {
	return nil
}
