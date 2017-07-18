package dispatcher

import (
	"github.com/bernos/ecso/pkg/ecso"
	"github.com/bernos/ecso/pkg/ecso/config"
)

// CommandFactory builds an ecso.Command instance using the options from config
type CommandFactory interface {
	Build(*config.Config) (ecso.Command, error)
}

// CommandFactoryFunc allows an ordinary func to be used as a CommandFactory interface
type CommandFactoryFunc func(*config.Config) (ecso.Command, error)

// Build calls fn
func (fn CommandFactoryFunc) Build(cfg *config.Config) (ecso.Command, error) {
	return fn(cfg)
}
