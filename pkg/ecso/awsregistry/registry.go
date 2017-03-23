package awsregistry

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

var (
	registries map[string]*registry = make(map[string]*registry)
)

type RegistryFactory interface {
	ForRegion(string) (Registry, error)
}

type RegistryFactoryFunc func(string) (Registry, error)

func (fn RegistryFactoryFunc) ForRegion(region string) (Registry, error) {
	return fn(region)
}

var DefaultRegistryFactory = RegistryFactoryFunc(func(region string) (Registry, error) {
	if registries[region] == nil {

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})

		if err != nil {
			return nil, fmt.Errorf("Failed to create AWSClientRegistry for region '%s': %s", region, err.Error())
		}

		registries[region] = NewRegistry(sess)
	}

	return registries[region], nil
})

type Registry interface {
	CloudFormationAPI() cloudformationiface.CloudFormationAPI
	CloudWatchLogsAPI() cloudwatchlogsiface.CloudWatchLogsAPI
	ECSAPI() ecsiface.ECSAPI
	Route53API() route53iface.Route53API
	S3API() s3iface.S3API
	SNSAPI() snsiface.SNSAPI
	STSAPI() stsiface.STSAPI
}

type registry struct {
	session *session.Session

	stsAPI            stsiface.STSAPI
	cloudFormationAPI cloudformationiface.CloudFormationAPI
	s3API             s3iface.S3API
	ecsAPI            ecsiface.ECSAPI
	cloudWatchLogsAPI cloudwatchlogsiface.CloudWatchLogsAPI
	route53           route53iface.Route53API
	snsAPI            snsiface.SNSAPI
}

func NewRegistry(sess *session.Session) *registry {
	return &registry{
		session: sess,
	}
}

func (r *registry) CloudFormationAPI() cloudformationiface.CloudFormationAPI {
	if r.cloudFormationAPI == nil {
		r.cloudFormationAPI = cloudformation.New(r.session)
	}
	return r.cloudFormationAPI
}

func (r *registry) CloudWatchLogsAPI() cloudwatchlogsiface.CloudWatchLogsAPI {
	if r.cloudWatchLogsAPI == nil {
		r.cloudWatchLogsAPI = cloudwatchlogs.New(r.session)
	}
	return r.cloudWatchLogsAPI
}

func (r *registry) ECSAPI() ecsiface.ECSAPI {
	if r.ecsAPI == nil {
		r.ecsAPI = ecs.New(r.session)
	}
	return r.ecsAPI
}

func (r *registry) Route53API() route53iface.Route53API {
	if r.route53 == nil {
		r.route53 = route53.New(r.session)
	}
	return r.route53
}

func (r *registry) S3API() s3iface.S3API {
	if r.s3API == nil {
		r.s3API = s3.New(r.session)
	}
	return r.s3API
}

func (r *registry) SNSAPI() snsiface.SNSAPI {
	if r.snsAPI == nil {
		r.snsAPI = sns.New(r.session)
	}
	return r.snsAPI
}

func (r *registry) STSAPI() stsiface.STSAPI {
	if r.stsAPI == nil {
		r.stsAPI = sts.New(r.session)
	}
	return r.stsAPI
}
