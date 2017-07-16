package config

import (
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso/api"
	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/ui"
)

type Config struct {
	Version string

	w      io.Writer
	reader io.Reader
	r      awsregistry.RegistryFactory

	serviceAPI     api.ServiceAPI
	environmentAPI api.EnvironmentAPI
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

func (c *Config) Reader() io.Reader {
	return c.reader
}

func (c *Config) ErrWriter() io.Writer {
	return ui.NewErrWriter(c.w)
}

func NewConfig(version string, options ...func(*Config)) (*Config, error) {
	cfg := &Config{
		Version: version,
		w:       os.Stderr,
		reader:  os.Stdin,
	}

	for _, o := range options {
		o(cfg)
	}

	return cfg, nil
}

func WithAWSRegistryFactory(r awsregistry.RegistryFactory) func(*Config) {
	return func(cfg *Config) {
		cfg.r = r
	}
}
