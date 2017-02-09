package api

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/helpers"
)

func (api *api) ServiceDown(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	log := api.cfg.Logger()
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	if err := api.deleteServiceStack(reg, env, service); err != nil {
		return err
	}

	log.Printf("\n")

	if err := api.clearServiceDNSRecords(reg, env, service); err != nil {
		return err
	}

	return nil
}

func (api *api) clearServiceDNSRecords(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log        = api.cfg.Logger()
		r53Service = helpers.NewRoute53Service(reg.Route53API(), log.PrefixPrintf("  "))
		dnsName    = fmt.Sprintf("%s.%s.", service.Name, env.CloudFormationParameters["DNSZone"])
	)

	log.Infof("Deleting any SRV DNS records for %s...", dnsName)

	if err := r53Service.DeleteResourceRecordSetsByName(dnsName, env.CloudFormationParameters["DNSZone"], "Deleted by ecso service down"); err != nil {
		return err
	}

	log.Printf("  Done\n")

	return nil
}

func (api *api) deleteServiceStack(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log   = api.cfg.Logger()
		stack = service.GetCloudFormationStackName(env)
		cfn   = helpers.NewCloudFormationService(env.Region, reg.CloudFormationAPI(), reg.S3API(), log.PrefixPrintf("  "))
	)

	log.Infof("Deleting cloud formation stack '%s'", stack)

	exists, err := cfn.StackExists(stack)

	if err != nil {
		return nil
	}

	if !exists {
		log.Printf("  Stack '%s' does not exist\n", stack)
		return nil
	}

	return cfn.DeleteStack(stack)
}
