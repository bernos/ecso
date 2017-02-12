package api

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/util"
)

func (api *api) DescribeEnvironment(env *ecso.Environment) (map[string]string, error) {
	var (
		log        = api.cfg.Logger()
		stack      = env.GetCloudFormationStackName()
		cfnConsole = util.CloudFormationConsoleURL(stack, env.Region)
		ecsConsole = util.ClusterConsoleURL(env.GetClusterName(), env.Region)
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationService(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.PrefixPrintf("  "))

	outputs, err := cfn.GetStackOutputs(stack)

	if err != nil {
		return nil, err
	}

	description := map[string]string{
		"Environment Name":     env.Name,
		"CloudFormation Stack": cfnConsole,
		"ECS Cluster":          ecsConsole,
		"Cluster Base URL":     fmt.Sprintf("http://%s", outputs["RecordSet"]),
	}

	for k, v := range outputs {
		description[k] = v
	}

	return description, nil
}
