package config

import (
	"io"
	"os"

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
