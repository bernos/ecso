package api

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
	"github.com/bernos/ecso/pkg/ecso/util"
)

type ServiceDescription struct {
	Name                     string
	URL                      string
	CloudFormationConsoleURL string
	CloudWatchLogsConsoleURL string
	ECSConsoleURL            string
	CloudFormationOutputs    map[string]string
}

func (api *api) DescribeService(env *ecso.Environment, service *ecso.Service) (*ServiceDescription, error) {
	var (
		log = api.cfg.Logger()
	)

	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return nil, err
	}

	cfn := helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())

	envOutputs, err := cfn.GetStackOutputs(env.GetCloudFormationStackName())

	if err != nil {
		return nil, err
	}

	serviceOutputs, err := cfn.GetStackOutputs(service.GetCloudFormationStackName(env))

	if err != nil {
		return nil, err
	}

	desc := &ServiceDescription{
		Name:                     service.Name,
		ECSConsoleURL:            util.ServiceConsoleURL(serviceOutputs["Service"], env.GetClusterName(), env.Region),
		CloudFormationConsoleURL: util.CloudFormationConsoleURL(service.GetCloudFormationStackName(env), env.Region),
		CloudWatchLogsConsoleURL: util.CloudWatchLogsConsoleURL(serviceOutputs["CloudWatchLogsGroup"], env.Region),
		CloudFormationOutputs:    make(map[string]string),
	}

	if service.Route != "" {
		desc.URL = fmt.Sprintf("http://%s%s", envOutputs["RecordSet"], service.Route)
	}

	for k, v := range serviceOutputs {
		desc.CloudFormationOutputs[k] = v
	}

	return desc, nil
}
