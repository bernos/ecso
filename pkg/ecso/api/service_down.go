package api

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/bernos/ecso/pkg/ecso"
)

func (api *api) ServiceDown(project *ecso.Project, env *ecso.Environment, service *ecso.Service) error {
	log := api.cfg.Logger
	reg, err := api.cfg.GetAWSClientRegistry(env.Region)

	if err != nil {
		return err
	}

	if err := api.stopAndDeleteECSService(reg, env, service); err != nil {
		return err
	}

	log.Printf("\n")

	if err := api.clearServiceDNSRecords(reg, env, service); err != nil {
		return err
	}

	log.Printf("\n")

	if err := api.deleteServiceStack(reg, env, service); err != nil {
		return err
	}

	return nil
}

func (api *api) ecsServiceExists(reg *ecso.AWSClientRegistry, service *ecso.Service, env *ecso.Environment) (bool, error) {
	ecsAPI := reg.ECSAPI()

	resp, err := ecsAPI.DescribeServices(&ecs.DescribeServicesInput{
		Cluster: aws.String(env.GetClusterName()),
		Services: []*string{
			aws.String(service.GetECSServiceName()),
		},
	})

	if err != nil {
		return false, err
	}

	return len(resp.Services) > 0, nil
}

func (api *api) stopAndDeleteECSService(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log    = api.cfg.Logger
		ecsAPI = reg.ECSAPI()
	)

	exists, err := api.ecsServiceExists(reg, service, env)

	if err != nil {
		return nil
	}

	if exists {
		log.Infof("Stopping ECS service '%s'", service.GetECSServiceName())

		if err := api.stopECSService(reg, env, service); err != nil {
			return err
		}

		log.Printf("\n")
		log.Infof("Deleting ECS service '%s'", service.GetECSServiceName())

		if _, err := ecsAPI.DeleteService(&ecs.DeleteServiceInput{
			Cluster: aws.String(env.GetClusterName()),
			Service: aws.String(service.GetECSServiceName()),
		}); err != nil {
			return err
		}

		log.Printf("  Done\n")

	} else {
		log.Infof("ECS service '%s' doesn't exists, skipping ecs teardown", service.GetECSServiceName())
	}

	return nil
}

func (api *api) stopECSService(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log        = api.cfg.Logger.PrefixPrintf("  ")
		ecsAPI     = reg.ECSAPI()
		ecsService = reg.ECSService(log)
	)

	describeServiceInput := &ecs.DescribeServicesInput{
		Cluster: aws.String(env.GetClusterName()),
		Services: []*string{
			aws.String(service.GetECSServiceName()),
		},
	}

	// First check if the service is running
	description, err := ecsAPI.DescribeServices(describeServiceInput)

	if err != nil {
		return err
	}

	// Nothing to do
	if len(description.Services) == 0 {
		log("No service named '%s' was found in the cluster '%s'", service.GetECSServiceName(), env.GetClusterName())
		return nil
	}

	if len(description.Services) > 1 {
		return fmt.Errorf("Found more than one ecs service named '%s'", service.GetECSServiceName())
	}

	status := *description.Services[0].Status

	if status == "ACTIVE" {
		log("Setting desired count to 0...\n")

		_, err = ecsAPI.UpdateService(&ecs.UpdateServiceInput{
			Cluster:      aws.String(env.GetClusterName()),
			Service:      aws.String(service.GetECSServiceName()),
			DesiredCount: aws.Int64(0),
		})

		if err != nil {
			return err
		}

		log("Waiting for ECS tasks to drain, and service to become stable...\n")

		cancel := ecsService.LogServiceEvents(service.GetECSServiceName(), env.GetClusterName(), func(e *ecs.ServiceEvent, err error) {
			if err == nil && e != nil {
				log("  %s %s\n", *e.CreatedAt, *e.Message)
			}
		})

		defer cancel()

		if err := ecsAPI.WaitUntilServicesStable(describeServiceInput); err != nil {
			return err
		}

		log("Deleting ECS service...\n")

		_, err = ecsAPI.DeleteService(&ecs.DeleteServiceInput{
			Cluster: aws.String(env.GetClusterName()),
			Service: aws.String(service.GetECSServiceName()),
		})

		if err != nil {
			return err
		}

		log("Waiting for ECS service to become inactive...\n")

		if err := ecsAPI.WaitUntilServicesInactive(describeServiceInput); err != nil {
			return err
		}
	} else if status == "DRAINING" {

		log("Waiting for ECS service to become inactive...\n")

		if err := ecsAPI.WaitUntilServicesInactive(describeServiceInput); err != nil {
			return err
		}
	} else if status == "INACTIVE" {
		log("ECS service is already inactive\n")
	}

	log("Successfully stopped ECS service\n")

	return nil

}

func (api *api) clearServiceDNSRecords(reg *ecso.AWSClientRegistry, env *ecso.Environment, service *ecso.Service) error {
	var (
		log        = api.cfg.Logger
		r53Service = reg.Route53Service(log.PrefixPrintf("  "))
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
		log   = api.cfg.Logger
		stack = service.GetCloudFormationStackName(env)
		cfn   = reg.CloudFormationService(log.PrefixPrintf("  "))
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
