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

func NewConfig(options ...func(*Config)) *Config {
	sess := session.New(&aws.Config{
		Region: aws.String("ap-southeast-2"),
	})

	cfg := &Config{
		Logger: NewLogger(os.Stdout),
		STS:    sts.New(sess),
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg
}
