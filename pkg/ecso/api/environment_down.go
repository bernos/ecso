package api

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
)

func (api *api) EnvironmentDown(p *ecso.Project, env *ecso.Environment) error {
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	var (
		log            = api.cfg.Logger()
		cfnHelper      = helpers.NewCloudFormationHelper(env.Region, reg.CloudFormationAPI(), reg.S3API(), reg.STSAPI(), log.Child())
		r53Helper      = helpers.NewRoute53Helper(reg.Route53API(), log.Child())
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
		datadogDNSName = fmt.Sprintf("%s.%s.%s", "datadog", env.GetClusterName(), zone)
	)

	for _, service := range p.Services {
		if err := api.ServiceDown(p, env, service); err != nil {
			return err
		}
		log.Printf("\n")
	}

	log.Infof("Deleting environment Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnHelper.DeleteStack(env.GetCloudFormationStackName()); err != nil {
		return err
	}

	log.Printf("\n")
	log.Infof("Deleting %s SRV records", datadogDNSName)

	return r53Helper.DeleteResourceRecordSetsByName(
		datadogDNSName,
		zone,
		"Deleted by ecso environment rm")
}
