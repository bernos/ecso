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
		cfnService     = helpers.NewCloudFormationService(env.Region, reg.CloudFormationAPI(), reg.S3API(), log.PrefixPrintf("  "))
		r53Service     = helpers.NewRoute53Service(reg.Route53API(), log.PrefixPrintf("  "))
		zone           = fmt.Sprintf("%s.", env.CloudFormationParameters["DNSZone"])
		datadogDNSName = fmt.Sprintf("%s.%s.%s", "datadog", env.GetClusterName(), zone)
	)

	for _, service := range p.Services {
		if err := api.ServiceDown(p, env, service); err != nil {
			return err
		}
	}

	log.Printf("\n")
	log.Infof("Deleting environment Cloud Formation stack '%s'", env.GetCloudFormationStackName())

	if err := cfnService.DeleteStack(env.GetCloudFormationStackName()); err != nil {
		return err
	}

	log.Printf("\n")
	log.Infof("Deleting %s SRV records", datadogDNSName)

	return r53Service.DeleteResourceRecordSetsByName(
		datadogDNSName,
		zone,
		"Deleted by ecso environment rm")
}
