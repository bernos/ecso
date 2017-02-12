package api

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/util"
)

func (api *api) DescribeService(env *ecso.Environment, service *ecso.Service) (map[string]string, error) {
	var (
		log = api.cfg.Logger()
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.PrefixPrintf("  "))

	envOutputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	serviceOutputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	items := map[string]string{
		"Service Console": util.ServiceConsoleURL(serviceOutputs["Service"], env.GetClusterName(), env.Region),
	}

	if service.Route != "" {
		items["Service URL"] = fmt.Sprintf("http://%s%s", envOutputs["RecordSet"], service.Route)
	}

	for k, v := range serviceOutputs {
		items[k] = v
	}

	return items, nil
}
