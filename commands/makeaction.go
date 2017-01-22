package commands

import (
	"github.com/bernos/ecso/pkg/ecso"
	"gopkg.in/urfave/cli.v1"
)

type factory func(*cli.Context) ecso.Command

// MakeAction is a factory func for generating wrapped ecso.Commands compatible
// with the urfave/cli command line interface semantics and types
func MakeAction(fn factory, dispatcher ecso.Dispatcher, options ...func(*ecso.DispatchOptions)) func(*cli.Context) error {
	return func(c *cli.Context) error {
		if err := dispatcher.Dispatch(fn(c), options...); err != nil {
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
