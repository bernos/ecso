package ecso

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"github.com/bernos/ecso/pkg/ecso/services"
)

type Config struct {
	Logger                Logger
	STS                   stsiface.STSAPI
	CloudFormationService services.CloudFormationService
}

func NewConfig(options ...func(*Config)) (*Config, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	})

	if err != nil {
		return nil, err
	}

	log := NewLogger(os.Stdout)

	cfn := services.NewCloudFormationService("ap-southeast-2", cloudformation.New(sess), s3.New(sess), log.PrefixPrintf("  "))

	cfg := &Config{
		Logger: log,
		STS:    sts.New(sess),
		CloudFormationService: cfn,
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}
