package commands

type Command interface {
	Execute() error
}

type CommandFunc func() error

func (fn CommandFunc) Execute() error {
	return fn()
}

func CommandError(err error) Command {
	return CommandFunc(func() error {
		return err
	})
}
