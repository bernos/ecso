package ecso

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type Config struct {
	Logger Logger
	STS    stsiface.STSAPI
}

func NewConfig(options ...func(*Config)) (*Config, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	})

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Logger: NewLogger(os.Stdout),
		STS:    sts.New(sess),
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}
