package ecso

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
	l Logger
	w io.Writer

	// AWS client registries by region
	awsClientRegistries map[string]*AWSClientRegistry
}

func (c *Config) Logger() Logger {
	if c.l == nil {
		c.l = NewLogger(c.w, "")
	}
	return c.l
}

func (c *Config) MustGetAWSClientRegistry(region string) *AWSClientRegistry {
	reg, err := c.GetAWSClientRegistry(region)

	if err != nil {
		c.Logger().Errorf("Failed to create AWSClientRegistry for region '%s': %s", region, err.Error())
		os.Exit(1)
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
	cfg := &Config{
		w:                   os.Stderr,
		awsClientRegistries: make(map[string]*AWSClientRegistry),
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}

func WriteOutputTo(w io.Writer) func(*Config) {
	return func(c *Config) {
		c.l = nil
		c.w = w
	}
}
