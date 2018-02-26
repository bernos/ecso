package ecso

import "fmt"

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

type EnvironmentExistsError struct {
	name string
}

func NewEnvironmentExistsError(name string) error {
	return &EnvironmentExistsError{name}
}

func (err *EnvironmentExistsError) Error() string {
	return fmt.Sprintf("An environment named '%s' already exists for this project.", err.name)
}

func IsArgumentRequiredError(err error) bool {
	_, ok := err.(*ArgumentRequiredError)
	return ok
}

func IsOptionRequiredError(err error) bool {
	_, ok := err.(*OptionRequiredError)
	return ok
}
