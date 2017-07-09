package config

import (
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/log"
	"github.com/bernos/ecso/pkg/ecso/ui"
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

func (c *Config) awsRegistryFactory() awsregistry.RegistryFactory {
	if c.r == nil {
		c.r = awsregistry.DefaultRegistryFactory
	}
	return c.r
}

func (c *Config) ServiceAPI() api.ServiceAPI {
	if c.serviceAPI == nil {
		c.serviceAPI = api.NewServiceAPI(c.w, c.awsRegistryFactory())
	}
	return c.serviceAPI
}

func (c *Config) EnvironmentAPI() api.EnvironmentAPI {
	if c.environmentAPI == nil {
		c.environmentAPI = api.NewEnvironmentAPI(c.w, c.awsRegistryFactory())
	}
	return c.environmentAPI
}

func (c *Config) Writer() io.Writer {
	return c.w
}

func (c *Config) ErrWriter() io.Writer {
	return ui.ErrWriter(c.w)
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
