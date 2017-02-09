package helpers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type Route53Service interface {
	DeleteResourceRecordSetsByName(name, zone, reason string) error
}

func NewRoute53Service(route53API route53iface.Route53API, log func(string, ...interface{})) Route53Service {
	return &route53Service{
		route53API: route53API,
		log:        log,
	}
}

type route53Service struct {
	route53API route53iface.Route53API
	log        func(string, ...interface{})
}

func (svc *route53Service) DeleteResourceRecordSetsByName(name, zone, reason string) error {
	zones, err := svc.route53API.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zone),
	})

	if err != nil {
		return err
	}

	for _, zone := range zones.HostedZones {
		resp, err := svc.route53API.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		})

		if err != nil {
			return err
		}

		changes := make([]*route53.Change, 0)

		for _, record := range resp.ResourceRecordSets {
			if *record.Name == name {
				changes = append(changes, &route53.Change{
					Action:            aws.String("DELETE"),
					ResourceRecordSet: record,
				})

				svc.log("Deleting recordset %s\n", *record.Name)
			}
		}

		if len(changes) > 0 {
			if _, err := svc.route53API.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
				HostedZoneId: zone.Id,
				ChangeBatch: &route53.ChangeBatch{
					Comment: aws.String(reason),
					Changes: changes,
				},
			}); err != nil {
				return err
			}

			svc.log("Done\n")
		} else {
			svc.log("No recordsets matching '%s' found\n", name)
		}
	}

	return nil
}
