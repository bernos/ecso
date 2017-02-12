package api

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
)

func (api *api) GetECSService(p *ecso.Project, env *ecso.Environment, s *ecso.Service) (*ecs.Service, error) {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	var (
		log    = api.cfg.Logger()
		cfn    = helpers.NewCloudFormationService(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.PrefixPrintf("  "))
		ecsAPI = reg.ECSAPI()
	)

	outputs, err := cfn.GetStackOutputs(s.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	if serviceName, ok := outputs["Service"]; ok {
		resp, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
			Cluster: aws.String(env.GetClusterName()),
			Services: []*string{
				aws.String(serviceName),
			},
		})

		if err != nil {
			return nil, err
		}

		if len(resp.Services) > 1 {
			return nil, fmt.Errorf("More than one service named '%s' was found", serviceName)
		}

		return resp.Services[0], nil
	}

	return nil, nil
}
