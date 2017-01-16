package commands

import "github.com/bernos/ecso/logger"

type Command interface {
	Execute(log logger.Logger) error
}

type CommandFunc func(logger.Logger) error

func (fn CommandFunc) Execute(log logger.Logger) error {
	return fn(log)
}

func CommandError(err error) Command {
	return CommandFunc(func(log logger.Logger) error {
		return err
	})
}
