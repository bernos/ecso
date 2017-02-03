package ecso

import (
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
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso/services"
)

type AWSClientRegistry struct {
	session *session.Session

	stsAPI            stsiface.STSAPI
	cloudFormationAPI cloudformationiface.CloudFormationAPI
	s3API             s3iface.S3API
	ecsAPI            ecsiface.ECSAPI
	cloudWatchLogsAPI cloudwatchlogsiface.CloudWatchLogsAPI
	route53           route53iface.Route53API

	cloudFormationService services.CloudFormationService
	ecsService            services.ECSService
	route53Service        services.Route53Service
}

func NewAWSClientRegistry(sess *session.Session) *AWSClientRegistry {
	return &AWSClientRegistry{
		session: sess,
	}
}

func (r *AWSClientRegistry) CloudFormationAPI() cloudformationiface.CloudFormationAPI {
	if r.cloudFormationAPI == nil {
		r.cloudFormationAPI = cloudformation.New(r.session)
	}
	return r.cloudFormationAPI
}

func (r *AWSClientRegistry) CloudWatchLogsAPI() cloudwatchlogsiface.CloudWatchLogsAPI {
	if r.cloudWatchLogsAPI == nil {
		r.cloudWatchLogsAPI = cloudwatchlogs.New(r.session)
	}
	return r.cloudWatchLogsAPI
}

func (r *AWSClientRegistry) ECSAPI() ecsiface.ECSAPI {
	if r.ecsAPI == nil {
		r.ecsAPI = ecs.New(r.session)
	}
	return r.ecsAPI
}

func (r *AWSClientRegistry) Route53API() route53iface.Route53API {
	if r.route53 == nil {
		r.route53 = route53.New(r.session)
	}
	return r.route53
}

func (r *AWSClientRegistry) S3API() s3iface.S3API {
	if r.s3API == nil {
		r.s3API = s3.New(r.session)
	}
	return r.s3API
}

func (r *AWSClientRegistry) STSAPI() stsiface.STSAPI {
	if r.stsAPI == nil {
		r.stsAPI = sts.New(r.session)
	}
	return r.stsAPI
}

func (r *AWSClientRegistry) CloudFormationService(log func(string, ...interface{})) services.CloudFormationService {
	if r.cloudFormationService == nil {
		r.cloudFormationService = services.NewCloudFormationService(*r.session.Config.Region, r.CloudFormationAPI(), r.S3API(), log)
	}
	return r.cloudFormationService
}

func (r *AWSClientRegistry) ECSService(log func(string, ...interface{})) services.ECSService {
	if r.ecsService == nil {
		r.ecsService = services.NewECSService(r.ECSAPI(), log)
	}
	return r.ecsService
}

func (r *AWSClientRegistry) Route53Service(log func(string, ...interface{})) services.Route53Service {
	if r.route53Service == nil {
		r.route53Service = services.NewRoute53Service(r.Route53API(), log)
	}
	return r.route53Service
}
