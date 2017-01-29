package commands

import (
	"fmt"

	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

type factory func(*cli.Context) (ecso.Command, error)

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(dispatcher ecso.Dispatcher, fn factory, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(c *cli.Context) error {
		command, err := fn(c)

		if err != nil {
			cli.ShowSubcommandHelp(c)
			return cli.NewExitError(err.Error(), 1)
		}

		if err := dispatcher.Dispatch(command, options...); err != nil {
			// if awsErr, ok := err.(awserr.Error); ok {
			// 	fmt.Printf("AWS ERROR\n%s", awsErr.Code())

			// 	if reqErr, ok := err.(awserr.RequestFailure); ok {
			// 		// A service error occurred
			// 		fmt.Println(reqErr.StatusCode(), reqErr.RequestID())
			// 	}
			// }

			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	}
}

type UsageError interface {
	Option() string
}

type OptionRequiredError struct {
	option string
}

func NewOptionRequiredError(option string) error {
	return &OptionRequiredError{option}
}

func (err *OptionRequiredError) Error() string {
	return fmt.Sprintf("--%s is a required option", err.option)
}

func (err *OptionRequiredError) Option() string {
	return err.option
}

type ArgumentRequiredError struct {
	arg string
}

func NewArgumentRequiredError(arg string) error {
	return &ArgumentRequiredError{arg}
}

func (err *ArgumentRequiredError) Error() string {
	return fmt.Sprintf("%s argument is required", err.arg)
}

func (err *ArgumentRequiredError) Option() string {
	return err.arg
}
