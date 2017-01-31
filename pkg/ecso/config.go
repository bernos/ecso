package ecso

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	Logger Logger

	// AWS client registries by region
	awsClientRegistries map[string]*AWSClientRegistry
}

func (c *Config) MustGetAWSClientRegistry(region string) *AWSClientRegistry {
	reg, err := c.GetAWSClientRegistry(region)

	if err != nil {
		c.Logger.Fatalf("Failed to create AWSClientRegistry for region '%s': %s", region, err.Error())
	}

	return reg
}

func (c *Config) GetAWSClientRegistry(region string) (*AWSClientRegistry, error) {
	if c.awsClientRegistries[region] == nil {

		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})

		if err != nil {
			return nil, err
		}

		c.awsClientRegistries[region] = NewAWSClientRegistry(sess)
	}

	return c.awsClientRegistries[region], nil
}

func NewConfig(options ...func(*Config)) (*Config, error) {
	log := NewLogger(os.Stdout)

	cfg := &Config{
		awsClientRegistries: make(map[string]*AWSClientRegistry),
		Logger:              log,
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}
