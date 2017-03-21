package awsregistry

import (
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
	registries map[string]*ClientRegistry = make(map[string]*ClientRegistry)
)

type ClientRegistry struct {
	session *session.Session

	stsAPI            stsiface.STSAPI
	cloudFormationAPI cloudformationiface.CloudFormationAPI
	s3API             s3iface.S3API
	ecsAPI            ecsiface.ECSAPI
	cloudWatchLogsAPI cloudwatchlogsiface.CloudWatchLogsAPI
	route53           route53iface.Route53API
	snsAPI            snsiface.SNSAPI
}

func ForRegion(region string) (*ClientRegistry, error) {
	if registries[region] == nil {

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})

		if err != nil {
			return nil, err
		}

		registries[region] = NewClientRegistry(sess)
	}

	return registries[region], nil
}

func NewClientRegistry(sess *session.Session) *ClientRegistry {
	return &ClientRegistry{
		session: sess,
	}
}

func (r *ClientRegistry) CloudFormationAPI() cloudformationiface.CloudFormationAPI {
	if r.cloudFormationAPI == nil {
		r.cloudFormationAPI = cloudformation.New(r.session)
	}
	return r.cloudFormationAPI
}

func (r *ClientRegistry) CloudWatchLogsAPI() cloudwatchlogsiface.CloudWatchLogsAPI {
	if r.cloudWatchLogsAPI == nil {
		r.cloudWatchLogsAPI = cloudwatchlogs.New(r.session)
	}
	return r.cloudWatchLogsAPI
}

func (r *ClientRegistry) ECSAPI() ecsiface.ECSAPI {
	if r.ecsAPI == nil {
		r.ecsAPI = ecs.New(r.session)
	}
	return r.ecsAPI
}

func (r *ClientRegistry) Route53API() route53iface.Route53API {
	if r.route53 == nil {
		r.route53 = route53.New(r.session)
	}
	return r.route53
}

func (r *ClientRegistry) S3API() s3iface.S3API {
	if r.s3API == nil {
		r.s3API = s3.New(r.session)
	}
	return r.s3API
}

func (r *ClientRegistry) SNSAPI() snsiface.SNSAPI {
	if r.snsAPI == nil {
		r.snsAPI = sns.New(r.session)
	}
	return r.snsAPI
}

func (r *ClientRegistry) STSAPI() stsiface.STSAPI {
	if r.stsAPI == nil {
		r.stsAPI = sts.New(r.session)
	}
	return r.stsAPI
}
