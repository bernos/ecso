package ui

import "fmt"

// StringValidator validates a string value
type StringValidator interface {
	Validate(string) error
}

// StringValidatorFunc is an adapter to allow the use of ordinary functions as
// string validators.
type StringValidatorFunc func(string) error

// Validate calls fn
func (fn StringValidatorFunc) Validate(s string) error {
	return fn(s)
}

// IntValidator validates an int value
type IntValidator interface {
	Validate(int) error
}

// IntValidatorFunc is an adaptor to allow the user of ordinary functions as
// in validators
type IntValidatorFunc func(int) error

// Validate calls fn
func (fn IntValidatorFunc) Validate(x int) error {
	return fn(x)
}

func ValidateIntBetween(min, max int) IntValidator {
	return IntValidatorFunc(func(v int) error {
		if v < min || v > max {
			return fmt.Errorf("A value between %d and %d is required", min, max)
		}
		return nil
	})
}

func ValidateAny() StringValidator {
	return StringValidatorFunc(func(v string) error {
		return nil
	})
}

func ValidateNotEmpty(msg string) StringValidator {
	return StringValidatorFunc(func(v string) error {
		if v == "" {
			return fmt.Errorf(msg)
		}
		return nil
	})
}

func ValidateRequired(name string) StringValidator {
	return ValidateNotEmpty(fmt.Sprintf("%s is required.", name))
}
