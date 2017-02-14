package ui

import "fmt"

func ValidateIntBetween(min, max int) func(int) error {
	return func(v int) error {
		if v < min || v > max {
			return fmt.Errorf("A value between %d and %d is required", min, max)
		}
		return nil
	}
}

func ValidateAny() func(string) error {
	return func(v string) error {
		return nil
	}
}

func ValidateNotEmpty(msg string) func(string) error {
	return func(v string) error {
		if v == "" {
			return fmt.Errorf(msg)
		}
		return nil
	}
}

func ValidateRequired(name string) func(string) error {
	return ValidateNotEmpty(fmt.Sprintf("%s is required.", name))
}
