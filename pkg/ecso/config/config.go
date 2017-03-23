package config

import (
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso/awsregistry"
	"github.com/bernos/ecso/pkg/ecso/log"
)

type Config struct {
	Version string

	l log.Logger
	w io.Writer
}

func (c *Config) Logger() log.Logger {
	if c.l == nil {
		c.l = log.NewLogger(c.w, "")
	}
	return c.l
}

func (c *Config) MustGetAWSClientRegistry(region string) *awsregistry.ClientRegistry {
	reg, err := awsregistry.ForRegion(region)

	if err != nil {
		c.Logger().Errorf("Failed to create AWSClientRegistry for region '%s': %s", region, err.Error())
		os.Exit(1)
	}

	return reg
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
