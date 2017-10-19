package mocks

import "github.com/aws/aws-sdk-go/service/route53/route53iface"

type Route53APIMock struct {
	route53iface.Route53API
}
