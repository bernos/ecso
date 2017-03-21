package ecso

import (
	"io"
	"os"

	"github.com/bernos/ecso/pkg/ecso/awsregistry"
)

type Config struct {
	Version string

	l Logger
	w io.Writer
}

func (c *Config) Logger() Logger {
	if c.l == nil {
		c.l = NewLogger(c.w, "")
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

func WriteOutputTo(w io.Writer) func(*Config) {
	return func(c *Config) {
		c.l = nil
		c.w = w
	}
}
