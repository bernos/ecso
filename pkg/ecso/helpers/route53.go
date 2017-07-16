package helpers

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
)

type Route53Helper interface {
	DeleteResourceRecordSetsByName(name, zone, reason string) error
}

func NewRoute53Helper(route53API route53iface.Route53API, w io.Writer) Route53Helper {
	return &route53Helper{
		w:          w,
		route53API: route53API,
	}
}

type route53Helper struct {
	w          io.Writer
	route53API route53iface.Route53API
}

func (h *route53Helper) DeleteResourceRecordSetsByName(name, zone, reason string) error {
	zones, err := h.route53API.ListHostedZonesByName(&route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zone),
	})

	if err != nil {
		return err
	}

	for _, zone := range zones.HostedZones {
		resp, err := h.route53API.ListResourceRecordSets(&route53.ListResourceRecordSetsInput{
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

				fmt.Fprintf(h.w, "Deleting recordset %s\n", *record.Name)
			}
		}

		if len(changes) > 0 {
			if _, err := h.route53API.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
				HostedZoneId: zone.Id,
				ChangeBatch: &route53.ChangeBatch{
					Comment: aws.String(reason),
					Changes: changes,
				},
			}); err != nil {
				return err
			}

			fmt.Fprint(h.w, "Done\n")
		} else {

			fmt.Fprintf(h.w, "No recordsets matching '%s' found\n", name)
		}
	}

	return nil
}
