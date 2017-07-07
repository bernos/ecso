package config

import (
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/log"
)

type Config struct {
	Version string

	l log.Logger
	w io.Writer
	r awsregistry.RegistryFactory

	serviceAPI     api.ServiceAPI
	environmentAPI api.EnvironmentAPI
}

func (c *Config) Logger() log.Logger {
	if c.l == nil {
		c.l = log.NewLogger(c.w, "")
	}
	return c.l
}

func (c *Config) AWSRegistryFactory() awsregistry.RegistryFactory {
	if c.r == nil {
		c.r = awsregistry.DefaultRegistryFactory
	}
	return c.r
}

func (c *Config) ServiceAPI() api.ServiceAPI {
	if c.serviceAPI == nil {
		c.serviceAPI = api.NewServiceAPI(c.Logger(), c.AWSRegistryFactory())
	}
	return c.serviceAPI
}

func (c *Config) EnvironmentAPI() api.EnvironmentAPI {
	if c.environmentAPI == nil {
		c.environmentAPI = api.NewEnvironmentAPI(c.Logger(), c.AWSRegistryFactory())
	}
	return c.environmentAPI
}

func NewConfig(version string, options ...func(*Config)) (*Config, error) {
	cfg := &Config{
		Version: version,
		w:       os.Stderr,
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}

func WithLogger(l log.Logger) func(*Config) {
	return func(cfg *Config) {
		cfg.l = l
	}
}

func WithAWSRegistryFactory(r awsregistry.RegistryFactory) func(*Config) {
	return func(cfg *Config) {
		cfg.r = r
	}
}
